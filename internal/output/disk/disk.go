package disk

import (
	"errors"
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

//Disk struct
type Disk struct {
	FilePath       string
	FilePermission string
	Permission     os.FileMode
	TypeCheck      bool
	Transactions   chan types.Transaction
	stopChan       chan struct{}
	logger         zap.Logger
	wg             sync.WaitGroup
}

//TransactionChan just return the channel of the transactions
func (d *Disk) TransactionChan() chan<- types.Transaction {
	return d.Transactions
}

//Init func Initialize the disk output plugin
func (d *Disk) Init(config types.Config, logger zap.Logger) error {
	if d.FilePath, d.TypeCheck = config["filepath"].(string); !d.TypeCheck {
		return errors.New("FilePath must be a string")
	}
	path.Clean(d.FilePath)
	if d.FilePermission, d.TypeCheck = config["permission"].(string); !d.TypeCheck {
		return errors.New("FilePermission must be from a string")
	}
	perm32, err := strconv.ParseUint(d.FilePermission, 0, 32)
	if err != nil {
		return err
	}
	d.Permission = os.FileMode(perm32)
	d.Transactions = make(chan types.Transaction)
	d.stopChan = make(chan struct{})
	d.logger = logger
	return nil
}

//WriteOnDisk func takes the transaction
//that to be written on the disk
func (d *Disk) writeOnDisk(transaction types.Transaction) {
	defer d.wg.Done()
	ack := true
	dir := filepath.Dir(d.FilePath)
	var err error
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
	}
	if err == nil {
		bytes, errRead := ioutil.ReadAll(transaction)
		err = errRead
		if err == nil {
			err = ioutil.WriteFile(d.FilePath, bytes, d.Permission)
		}
	}
	if err != nil {
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
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			select {
			case <-d.stopChan:
				return
			case transaction, _ := <-d.Transactions:
				d.wg.Add(1)
				go d.writeOnDisk(transaction)
			}
		}
	}()

	return nil
}

//Close func Send a close signal to stop chan
// to stop taking transactions and Close everything safely
func (d *Disk) Close(time.Duration) error {
	d.stopChan <- struct{}{}
	d.wg.Wait()
	close(d.Transactions)
	close(d.stopChan)
	return nil
}
