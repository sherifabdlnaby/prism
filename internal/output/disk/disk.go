package disk

import (
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

type Disk struct {
	FileName       string
	FilePath       string
	Transactions   chan types.Transaction
	stopChan       chan struct{}
	logger         zap.Logger
	wg             sync.WaitGroup
	FilePermission os.FileMode
}

func (d *Disk) TransactionChan() chan<- types.Transaction {
	return d.Transactions
}

func (d *Disk) Init(config types.Config, logger zap.Logger) error {
	d.FileName = config["filename"].(string)
	d.FilePath = config["filepath"].(string)
	path.Clean(d.FilePath)
	d.FilePermission = config["permission"].(os.FileMode)
	d.Transactions = make(chan types.Transaction, 1)
	d.stopChan = make(chan struct{})
	d.logger = logger
	return nil
}

func (d *Disk) Start() error {
	d.logger.Info("Started Disk Output!")

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			select {
			case <-d.stopChan:
				d.logger.Info("Closing Disk Output...")
				return
			case transaction, more := <-d.Transactions:
				if more {
					d.wg.Add(1)
					go func() {
						defer d.wg.Done()
						d.logger.Info("RECEIVED OUTPUT TRANSACTION...")
						bytes, _ := ioutil.ReadAll(transaction)
						var err error = nil
						if _, err = os.Stat(d.FilePath); os.IsNotExist(err) {
							err = os.MkdirAll(d.FilePath, os.ModePerm)
						}
						if err == nil {
							FilePath := filepath.Join(d.FilePath, d.FileName)
							err = ioutil.WriteFile(FilePath, bytes, d.FilePermission)
						}
						if err != nil {
							d.logger.Info("Error in output: ", zap.Error(err))
							// send response
							transaction.ResponseChan <- types.Response{
								Error: err,
								Ack:   false,
							}
							return
						}

						d.logger.Info("OUTPUT SUCCESSFUL, Sending Response. ")

						// send response
						transaction.ResponseChan <- types.Response{
							Error: nil,
							Ack:   true,
						}

						return
					}()
				}
			}
		}
	}()

	return nil
}

func (d *Disk) Close(time.Duration) error {
	d.logger.Info("Sending closing signal to Disk Output...")
	d.stopChan <- struct{}{}
	d.wg.Wait()
	close(d.Transactions)
	d.logger.Info("Closed Disk Output.")
	return nil
}
