package input

import (
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"os"
	"time"
)

// Dummy Input that read a file from root just for testing.
type Dummy struct {
	FileName     string
	Transactions chan types.StreamableTransaction
	stopChan     chan struct{}
	logger       zap.Logger
}

func (d *Dummy) TransactionChan() <-chan types.StreamableTransaction {
	return d.Transactions
}

func (d *Dummy) Init(config types.Config, logger zap.Logger) error {
	d.FileName = config["filename"].(string)
	d.Transactions = make(chan types.StreamableTransaction, 1)
	d.stopChan = make(chan struct{})
	d.logger = logger
	return nil
}

func (d *Dummy) Start() error {
	d.logger.Info("Started Input, Hooray!")

	go func() {
		for {
			select {
			case <-d.stopChan:
				d.logger.Info("Closing...")
				break
			default:
				d.logger.Info("SENDING A TRANSACTION...")
				reader, err := os.Open(d.FileName)
				responseChan := make(chan types.Response)

				if err != nil {
					d.logger.Info("Error in input: ", zap.Error(err))
					continue
				}

				// Send Transaction
				d.Transactions <- types.StreamableTransaction{
					StreamablePayload: types.StreamablePayload{
						Name:      "test",
						Reader:    reader,
						ImageData: nil,
					},
					ResponseChan: responseChan,
				}

				// Wait Transaction
				response := <-responseChan

				d.logger.Info("RECEIVED RESPONSE.", zap.Any("response", response))

				time.Sleep(time.Second * 3)
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
