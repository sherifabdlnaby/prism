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

//Disk struct
type Disk struct {
	FileName       string
	FilePath       string
	Transactions   chan types.Transaction
	stopChan       chan struct{}
	logger         zap.Logger
	wg             sync.WaitGroup
	FilePermission os.FileMode
}

//TransactionChan just return the channel of the transactions
func (d *Disk) TransactionChan() chan<- types.Transaction {
	return d.Transactions
}

//Init func Initialize the disk output plugin
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

//WriteOnDisk func takes the disk and the transaction
//that to be written on the disk
func WriteOnDisk(d *Disk, transaction types.Transaction) {
	defer d.wg.Done()
	var err error
	ack := true
	d.logger.Info("RECEIVED OUTPUT TRANSACTION...")
	bytes, _ := ioutil.ReadAll(transaction)
	if _, err = os.Stat(d.FilePath); os.IsNotExist(err) {
		err = os.MkdirAll(d.FilePath, os.ModePerm)
	}
	if err == nil {
		FilePath := filepath.Join(d.FilePath, d.FileName)
		err = ioutil.WriteFile(FilePath, bytes, d.FilePermission)
		d.logger.Info("OUTPUT SUCCESSFUL, Sending Response. ")
	} else {
		d.logger.Info("Error in output: ", zap.Error(err))
		ack = false
	}
	// send response
	transaction.ResponseChan <- types.Response{
		Error: err,
		Ack:   ack,
	}
	return
}

// Start the plugin and be ready for taking transactions
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
					go WriteOnDisk(d, transaction)
				}
			}
		}
	}()

	return nil
}

//Close func Send a close signal to stop chan
// to stop taking transactions and Close everything safely
func (d *Disk) Close(time.Duration) error {
	d.logger.Info("Sending closing signal to Disk Output...")
	d.stopChan <- struct{}{}
	d.wg.Wait()
	close(d.Transactions)
	close(d.stopChan)
	d.logger.Info("Closed Disk Output.")
	return nil
}
