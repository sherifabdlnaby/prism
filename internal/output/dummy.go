package output

import (
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"io/ioutil"
	"time"
)

// Dummy Input that read a file from root just for testing.
type Dummy struct {
	FileName     string
	Transactions chan types.Transaction
	stopChan     chan struct{}
	logger       zap.Logger
}

func (d *Dummy) TransactionChan() chan<- types.Transaction {
	return d.Transactions
}

func (d *Dummy) Init(config types.Config, logger zap.Logger) error {
	d.FileName = config["filename"].(string)
	d.Transactions = make(chan types.Transaction, 1)
	d.stopChan = make(chan struct{})
	d.logger = logger
	return nil
}

func (d *Dummy) Start() error {
	d.logger.Info("Started Output, Hooray!")

	go func() {
		for {
			select {
			case <-d.stopChan:
				d.logger.Info("Closing...")
				break
			case transaction := <-d.Transactions:
				d.logger.Info("RECEIVED OUTPUT TRANSACTION...")

				err := ioutil.WriteFile(d.FileName, transaction.ImageBytes, 0644)

				if err != nil {
					d.logger.Info("Error in output: ", zap.Error(err))
					continue
				}

				d.logger.Info("OUTPUT SUCCESSFUL, Sending Response. ")

				// send response
				transaction.ResponseChan <- types.Response{
					Error: nil,
					Ack:   true,
				}
			}
		}
	}()

	return nil
}

func (d *Dummy) Close(time.Duration) error {
	d.logger.Info("Sending closing signal...")
	d.stopChan <- struct{}{}
	close(d.Transactions)
	d.logger.Info("Closed.")
	return nil
}
