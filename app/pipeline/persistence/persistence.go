package persistence

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"go.uber.org/zap"
)

const dbName = "async.db"

type Repository struct {
	db      *bolt.DB
	dbDir   string
	dataDir string
	logger  zap.SugaredLogger
}

func NewRepository(directory string, logger zap.SugaredLogger) (*Repository, error) {

	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return nil, err
	}
	dbDir := directory + "/" + dbName

	db, err := bolt.Open(dbDir, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("error while opening persistence Repository file %s", err.Error())
	}

	dataDir := directory + "/pipelines"
	err = os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:      db,
		dbDir:   dbDir,
		dataDir: dataDir,
		logger:  *logger.Named("persistence"),
	}, nil
}

type Bucket struct {
	db                *bolt.DB
	bucket, imagesDir string
	onInitFiles       []os.FileInfo
	logger            zap.SugaredLogger
}

func (r *Repository) Bucket(name, hash string, logger zap.SugaredLogger) (*Bucket, error) {

	bucketName := name + hash
	bucketImagesDir := r.dataDir + "/" + bucketName

	// make dir anyway
	_ = os.Mkdir(bucketImagesDir, os.ModePerm)

	// save current files (used in cleanup)
	existingFiles, err := ioutil.ReadDir(bucketImagesDir)
	if err != nil {
		return nil, err
	}

	b := &Bucket{
		db:          r.db,
		bucket:      bucketName,
		imagesDir:   bucketImagesDir,
		onInitFiles: existingFiles,
		logger:      *logger.Named(name),
	}

	err = b.init()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Bucket) init() error {
	return b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(b.bucket))
		if err != nil && err != bolt.ErrBucketExists {
			return fmt.Errorf("creating bucket: %s", err)
		}
		if err == bolt.ErrBucketExists {
			return nil
		}
		b.logger.Infof("new bucket created for pipeline %s", b.bucket)
		return nil
	})
}

func (b *Bucket) Cleanup() {

	// Here we remove any files we read before applying the requests,
	// this for the narrow possibility that a system crash happened after removing it from Repository but before deleting it from the file,
	// as the Repository is the source of truth, any remaining file after applying everything shall be removed.
	// Need to check if they're not removed After we read them as applying itself could have removed its own files.
	// Bottom line this function remove any image that has no entry in the Repository.
	for _, file := range b.onInitFiles {
		if !file.IsDir() {
			// get abs Path
			filePath := b.imagesDir + "/" + file.Name()
			filePath, _ = filepath.Abs(filePath)
			ext := filepath.Ext(filePath)
			if _, err := os.Stat(filePath); !os.IsNotExist(err) && ext == ".pri" {
				err := os.Remove(filePath)
				if err != nil {
					b.logger.Warn("an error occurred on directory cleanup")
				}
			}
		}
	}

}

func (b *Bucket) GetAllAsyncJobs() ([]job.Async, error) {
	JobList := make([]job.Async, 0)

	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.bucket))
		err := b.ForEach(func(k, v []byte) error {
			asyncJob := &job.Async{}
			err := json.Unmarshal(v, asyncJob)
			if err != nil {
				return err
			}

			err = asyncJob.Load(nil)
			if err != nil {
				return err
			}

			JobList = append(JobList, *asyncJob)
			return nil
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	return JobList, nil
}

func (b *Bucket) DeleteAsyncJob(asyncJob *job.Async) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.bucket))
		err := b.Delete([]byte(asyncJob.ID))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Delete from filesystem
	err = os.Remove(asyncJob.Filepath)
	if err != nil {
		return fmt.Errorf("failed to remove tmpFile after finalzing async job, error: %s", err.Error())
	}

	return nil
}

func (b *Bucket) CreateAsyncJob(node string, t job.Job) (*job.Async, error) {

	// --------------------- Write To Temp File -------------------------------------
	filepath, newPayload, err := b.writeToTmpFile(t.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to save async job to tmp file, error: %s", err.Error())
	}

	// --------------------- Create Async Job ---------------------------------------

	// Create Async Job
	asyncJob := &job.Async{
		ID:       uuid.New().String(),
		Node:     node,
		Filepath: filepath,
		Data:     t.Data,
	}

	err = asyncJob.Load(newPayload)
	if err != nil {
		return nil, err
	}

	// --------------------- Save to Repository ---------------------------------------------

	encodedBytes, err := json.Marshal(asyncJob)
	if err != nil {
		return nil, err
	}

	// Persist to Database
	err = b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.bucket))
		err := b.Put([]byte(asyncJob.ID), encodedBytes)
		return err
	})

	if err != nil {
		return nil, err
	}

	return asyncJob, nil
}

func (b *Bucket) writeToTmpFile(Payload payload.Payload) (filepath string, newPayload payload.Payload, err error) {

	tmpFile, err := ioutil.TempFile(b.imagesDir, "*.pri")
	if err != nil {
		return "", nil, err
	}
	filepath = tmpFile.Name()

	// Write Data to Disk
	switch Payload := Payload.(type) {
	case payload.Bytes:
		// Write Data
		nBytes, err := tmpFile.Write(Payload)

		if err != nil {
			return "", nil, err
		}

		if nBytes != len(Payload) {
			return "", nil, err
		}

		return filepath, Payload, err
	case payload.Stream:
		// Drain Stream Into File
		_, err = io.Copy(tmpFile, Payload)
		if err != nil {
			return "", nil, err
		}

		// Get to start of the file
		_, err = tmpFile.Seek(0, 0)
		if err != nil {
			return "", nil, err
		}

		// File Reader for input
		newPayload = payload.Stream(tmpFile)

		return filepath, newPayload, err
	default:
		return "", nil, fmt.Errorf("invalid job Payload type, must be Payload.Bytes or Payload.Stream")
	}
}
