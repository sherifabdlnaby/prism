package dummy

import (
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"os"
	"sync"
	"time"
)

// Dummy Input that read a file from root just for testing.
type Dummy struct {
	FileName     string
	Transactions chan types.Transaction
	stopChan     chan struct{}
	logger       zap.Logger
	wg           sync.WaitGroup
}

func (d *Dummy) TransactionChan() <-chan types.Transaction {
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
	d.logger.Info("Started Input, Hooray!")

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			select {
			case <-d.stopChan:
				d.logger.Info("Closing...")
				return
			default:
				go func() {
					d.logger.Info("SENDING A TRANSACTION...")
					reader, err := os.Open(d.FileName)
					responseChan := make(chan types.Response)

					if err != nil {
						d.logger.Info("Error in dummy: ", zap.Error(err))
						return
					}

					// Send Transaction
					d.Transactions <- types.Transaction{
						Payload: types.Payload{
							Name:      "test",
							Reader:    reader,
							ImageData: nil,
						},
						ResponseChan: responseChan,
					}

					// Wait Transaction
					response := <-responseChan

					d.logger.Info("RECEIVED RESPONSE.", zap.Any("response", response))
				}()
				time.Sleep(time.Millisecond * 500)
			}
		}
	}()

	return nil
}

func (d *Dummy) Close(time.Duration) error {
	d.logger.Info("Sending closing signal...")
	d.stopChan <- struct{}{}
	d.wg.Wait()
	close(d.Transactions)
	d.logger.Info("Closed.")
	return nil
}
