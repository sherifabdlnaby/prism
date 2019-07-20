package persistence

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"os"
	"time"
)

var db *bolt.DB
var initialized bool

func initializePersistence() error {
	if !initialized {
		dataDir := config.EnvPrismDataDir.Lookup()

		// open pipeline DB
		err := os.MkdirAll(dataDir, os.ModePerm)
		if err != nil {
			return err
		}

		err = os.MkdirAll(config.EnvPrismTmpDir.Lookup(), os.ModePerm)
		if err != nil {
			return err
		}

		db, err = bolt.Open(dataDir+"/async.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			return fmt.Errorf("error while opening persistence DB file %s", err.Error())
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

func (p *Persistence) GetAllTxn() ([]transaction.Async, error) {
	TxnList := make([]transaction.Async, 0)

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(p.bucket))
		err := b.ForEach(func(k, v []byte) error {
			asyncTxn := &transaction.Async{}
			err := json.Unmarshal(v, asyncTxn)
			if err != nil {
				return err
			}
			// open tmp file
			tmpFile, err := os.Open(asyncTxn.Filepath)
			if err != nil {
				p.logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
			}

			asyncTxn.TmpFile = tmpFile
			TxnList = append(TxnList, *asyncTxn)
			return nil
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	return TxnList, nil
}

func (p *Persistence) DeleteTxn(asyncTxn *transaction.Async) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(p.bucket))
		err := b.Delete([]byte(asyncTxn.ID))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Delete from filesystem
	err = os.Remove(asyncTxn.Filepath)
	if err != nil {
		return fmt.Errorf("failed to remove tmpFile after finalzing async transaction, error: %s", err.Error())
	}

	return nil
}

func (p *Persistence) SaveTxn(node string, t transaction.Transaction) (*transaction.Async, payload.Payload, error) {

	// --------------------- Write To Temp File -------------------------------------

	filepath, newPayload, err := writeToTmpFile(t.Payload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to save async transaction to tmp file, error: %s", err.Error())
	}

	// --------------------- Create Async Txn ---------------------------------------

	// Check if newPayload is a file, as so it must be saved with asyncTxn to be closed when it's done.
	tmpFile, ok := newPayload.(*os.File)
	if !ok {
		tmpFile = nil
	}

	// Create Async Txn
	asyncTxn := &transaction.Async{
		ID:       uuid.New().String(),
		Node:     node,
		Filepath: filepath,
		Data:     t.Data,
		TmpFile:  tmpFile,
	}

	// --------------------- Save to DB ---------------------------------------------

	encodedBytes, err := json.Marshal(asyncTxn)
	if err != nil {
		return nil, nil, err
	}

	// Persist to Database
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(p.bucket))
		err := b.Put([]byte(asyncTxn.ID), encodedBytes)
		return err
	})

	if err != nil {
		return nil, nil, err
	}

	return asyncTxn, newPayload, nil
}

func writeToTmpFile(Payload payload.Payload) (filepath string, newPayload payload.Payload, err error) {

	dirPath := config.EnvPrismTmpDir.Lookup()
	tmpFile, err := ioutil.TempFile(dirPath, "*.prism")
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
		return "", nil, fmt.Errorf("invalid transaction Payload type, must be Payload.Bytes or Payload.Stream")
	}
}
