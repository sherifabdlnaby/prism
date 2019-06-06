package disk

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

//Disk struct
type Disk struct {
	FilePermission     config.Value
	FilePath           config.Value
	Permission         os.FileMode
	TypeCheck          bool
	Transactions       <-chan transaction.Transaction
	StreamTransactions <-chan transaction.Streamable
	stopChan           chan struct{}
	logger             zap.SugaredLogger
	wg                 sync.WaitGroup
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &Disk{}
}

//TransactionChan just return the channel of the transactions
func (d *Disk) SetTransactionChan(t <-chan transaction.Transaction) {
	d.Transactions = t
}

//TransactionChan just return the channel of the transactions
func (d *Disk) SetStreamTransactionChan(t <-chan transaction.Streamable) {
	d.StreamTransactions = t
}

//Init func Initialize the disk output plugin
func (d *Disk) Init(config config.Config, logger zap.SugaredLogger) error {
	var err error

	d.FilePath, err = config.Get("filepath", nil)
	if err != nil {
		return err
	}

	d.FilePermission, err = config.Get("permission", nil)
	if err != nil {
		return err
	}

	perm32, err := strconv.ParseUint(d.FilePermission.Get().String(), 0, 32)

	if err != nil {
		return err
	}

	d.Permission = os.FileMode(perm32)
	d.stopChan = make(chan struct{})
	d.logger = logger

	return nil
}

//WriteOnDisk func takes the transaction
//that to be written on the disk
func (d *Disk) writeOnDisk(txn transaction.Transaction) {
	defer d.wg.Done()
	ack := true

	filePathV, err := d.FilePath.Evaluate(txn.Data)
	if err != nil {
		txn.ResponseChan <- response.Error(err)
		return
	}

	filePath := filePathV.String()
	dir := filepath.Dir(filePath)

	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
	}

	if err == nil {
		err = ioutil.WriteFile(filePath, txn.Payload, d.Permission)
	}

	if err != nil {
		ack = false
	}

	// send response
	txn.ResponseChan <- response.Response{
		Error: err,
		Ack:   ack,
	}
}

//WriteOnDisk func takes the transaction
//that to be written on the disk
func (d *Disk) writeOnDiskStream(txn transaction.Streamable) {
	defer d.wg.Done()
	ack := true

	filePathV, err := d.FilePath.Evaluate(txn.Data)
	if err != nil {
		txn.ResponseChan <- response.Error(err)
		return
	}

	filePath := filePathV.String()
	dir := filepath.Dir(filePath)

	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
	}

	if err == nil {
		bytes, errRead := ioutil.ReadAll(txn.Payload)
		err = errRead
		if err == nil {
			err = ioutil.WriteFile(filePath, bytes, d.Permission)
		}
	}

	if err != nil {
		ack = false
	}

	// send response
	txn.ResponseChan <- response.Response{
		Error: err,
		Ack:   ack,
	}
}

// Start the plugin and be ready for taking transactions
func (d *Disk) Start() error {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for txn := range d.Transactions {
			d.wg.Add(1)
			go d.writeOnDisk(txn)
		}
	}()

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for txn := range d.StreamTransactions {
			d.wg.Add(1)
			go d.writeOnDiskStream(txn)
		}
	}()

	return nil
}

//Close func Send a close signal to stop chan
// to stop taking transactions and Close everything safely
func (d *Disk) Close() error {
	d.wg.Wait()
	return nil
}
