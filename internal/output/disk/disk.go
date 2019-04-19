package disk

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

//Disk struct
type Disk struct {
	FilePermission string
	Permission     os.FileMode
	TypeCheck      bool
	Transactions   chan component.Transaction
	stopChan       chan struct{}
	logger         zap.SugaredLogger
	wg             sync.WaitGroup
	config         component.Config
}

func NewComponent() component.Component {
	return &Disk{}
}

//TransactionChan just return the channel of the transactions
func (d *Disk) TransactionChan() chan<- component.Transaction {
	return d.Transactions
}

//Init func Initialize the disk output plugin
func (d *Disk) Init(config component.Config, logger zap.SugaredLogger) error {
	d.config = config
	if !d.config.IsSet("filepath") {
		return fmt.Errorf("field [%s] is not set", "filepath")
	}

	FilePermission, err := d.config.Get("permission", nil)
	if err != nil {
		return err
	}

	d.FilePermission = FilePermission.String()
	perm32, err := strconv.ParseUint(d.FilePermission, 0, 32)

	if err != nil {
		return err
	}

	d.Permission = os.FileMode(perm32)
	d.Transactions = make(chan component.Transaction)
	d.stopChan = make(chan struct{})
	d.logger = logger

	return nil
}

//WriteOnDisk func takes the transaction
//that to be written on the disk
func (d *Disk) writeOnDisk(transaction component.Transaction) {
	defer d.wg.Done()
	ack := true

	filePathV, err := d.config.Get("filepath", transaction.ImageData)
	if err != nil {
		transaction.ResponseChan <- component.ResponseError(err)
		return
	}

	filePath := filePathV.String()
	dir := filepath.Dir(filePath)

	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
	}

	if err == nil {
		bytes, errRead := ioutil.ReadAll(transaction)
		err = errRead
		if err == nil {
			err = ioutil.WriteFile(filePath, bytes, d.Permission)
		}
	}

	if err != nil {
		ack = false
	}

	// send response
	transaction.ResponseChan <- component.Response{
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
