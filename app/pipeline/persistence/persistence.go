package persistence

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var db *bolt.DB
var initialized bool
var dbDir string
var dataDir string
var existingFiles []os.FileInfo

func initializePersistence() error {
	if !initialized {
		dbDir = config.EnvPrismDataDir.Lookup()
		dataDir = dbDir + "/images"

		err := os.MkdirAll(dataDir, os.ModePerm)
		if err != nil {
			return err
		}

		err = os.MkdirAll(dbDir, os.ModePerm)
		if err != nil {
			return err
		}

		db, err = bolt.Open(dbDir+"/async.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			return fmt.Errorf("error while opening persistence DB file %s", err.Error())
		}

		//save current files
		existingFiles, err = ioutil.ReadDir(dataDir)
		if err != nil {
			return err
		}

		initialized = true
	}
	return nil
}

type Persistence struct {
	bucket string
	logger zap.SugaredLogger
}

func NewPersistence(name, hash string, logger zap.SugaredLogger) (Persistence, error) {
	err := initializePersistence()
	if err != nil {
		return Persistence{}, err
	}

	bucketName := name + hash
	per := Persistence{bucket: bucketName, logger: *logger.Named("persistence")}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte(bucketName))
		if err != nil && err != bolt.ErrBucketExists {
			return fmt.Errorf("create bucket: %s", err)
		}
		if err == bolt.ErrBucketExists {
			return nil
		}
		per.logger.Infof("new pipeline, created a persistence bucket for new pipeline %s (%s)", name, hash)
		return nil
	})
	if err != nil {
		return Persistence{}, err
	}

	return per, nil
}

func DirectoryCleanup() {

	// Here we remove any files we read before applying the requests,
	// this for the narrow possibility that a system crash happened after removing it from DB but before deleting it from the file,
	// as the DB is the source of truth, any remaining file after applying everything shall be removed.
	// Need to check if they're not removed After we read them as applying itself could have removed its own files.
	// Bottom line this function remove any image that has no entry in the DB.
	for _, file := range existingFiles {
		if !file.IsDir() {
			// get abs Path
			filePath := dataDir + "/" + file.Name()
			filePath, _ = filepath.Abs(filePath)
			ext := filepath.Ext(filePath)
			if _, err := os.Stat(filePath); !os.IsNotExist(err) && ext == ".pri" {
				err := os.Remove(filePath)
				if err != nil {
					//TODO log here
				}
			}
		}
	}

}

func (p *Persistence) GetAllJobs() ([]job.Async, error) {
	JobList := make([]job.Async, 0)

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(p.bucket))
		err := b.ForEach(func(k, v []byte) error {
			asyncJob := &job.Async{}
			err := json.Unmarshal(v, asyncJob)
			if err != nil {
				return err
			}
			// open tmp file
			tmpFile, err := os.Open(asyncJob.Filepath)
			if err != nil {
				p.logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
			}

			asyncJob.TmpFile = tmpFile
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

func (p *Persistence) DeleteJob(asyncJob *job.Async) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(p.bucket))
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

func (p *Persistence) SaveJob(node string, t job.Job) (*job.Async, payload.Payload, error) {

	// --------------------- Write To Temp File -------------------------------------

	filepath, newPayload, err := writeToTmpFile(t.Payload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to save async job to tmp file, error: %s", err.Error())
	}

	// --------------------- Create Async Job ---------------------------------------

	// TODO USE FINALIZER

	// Check if newPayload is a file, as so it must be closed when it's done.
	tmpFile, ok := newPayload.(*os.File)
	if !ok {
		tmpFile = nil
	}

	// Create Async Job
	asyncJob := &job.Async{
		ID:       uuid.New().String(),
		Node:     node,
		Filepath: filepath,
		Data:     t.Data,
		TmpFile:  tmpFile,
	}

	// --------------------- Save to DB ---------------------------------------------

	encodedBytes, err := json.Marshal(asyncJob)
	if err != nil {
		return nil, nil, err
	}

	// Persist to Database
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(p.bucket))
		err := b.Put([]byte(asyncJob.ID), encodedBytes)
		return err
	})

	if err != nil {
		return nil, nil, err
	}

	return asyncJob, newPayload, nil
}

func writeToTmpFile(Payload payload.Payload) (filepath string, newPayload payload.Payload, err error) {

	tmpFile, err := ioutil.TempFile(dataDir, "*.pri")
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

		err = tmpFile.Close()
		if err != nil {
			return "", nil, err
		}

		return filepath, Payload, err
	case payload.Stream:
		// Drain Stream Into File
		_, err = io.Copy(tmpFile, Payload)
		if err != nil {
			_ = tmpFile.Close()
			return "", nil, err
		}

		// Get to start of the file
		_, err = tmpFile.Seek(0, 0)
		if err != nil {
			_ = tmpFile.Close()
			return "", nil, err
		}

		// File Reader for input, (REMEMBER, The file need to be closed by whoever using this)
		newPayload = payload.Stream(tmpFile)

		return filepath, newPayload, err
	default:
		return "", nil, fmt.Errorf("invalid job Payload type, must be Payload.Bytes or Payload.Stream")
	}
}
