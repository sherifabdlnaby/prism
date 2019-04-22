package dummy

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
	"os"
	"sync"
	"time"
)

// Dummy Input that read a file from root just for testing.
type Dummy struct {
	FileName     config.Value
	Pipeline     config.Value
	Transactions chan transaction.InputTransaction
	stopChan     chan struct{}
	logger       zap.SugaredLogger
	wg           sync.WaitGroup
	metric       int
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &Dummy{}
}

// TransactionChan Return Transaction Chan used to send transaction to this Component
func (d *Dummy) TransactionChan() <-chan transaction.InputTransaction {
	return d.Transactions
}

// Init Initializes Plugin
func (d *Dummy) Init(config config.Config, logger zap.SugaredLogger) error {
	FileName, err := config.Get("filename", nil)
	if err != nil {
		return err
	}

	Pipeline, err := config.Get("pipeline", nil)
	if err != nil {
		return err
	}

	d.FileName = FileName
	d.Pipeline = Pipeline

	d.Transactions = make(chan transaction.InputTransaction, 1)
	d.stopChan = make(chan struct{})
	d.logger = logger
	return nil
}

// Start Starts Plugin
func (d *Dummy) Start() error {
	d.logger.Debugw("Started Input, Hooray!")

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			select {
			case <-d.stopChan:
				d.logger.Debugw("Closing...")
				return
			default:

				go func(i int) {
					d.logger.Debugw("SENDING A TRANSACTION...")
					filename, _ := d.FileName.Evaluate(nil)

					reader, err := os.Open(filename.String())
					responseChan := make(chan transaction.Response)
					if err != nil {
						d.logger.Debugw("Error in dummy: ", zap.Error(err))
						return
					}

					pipeline, _ := d.Pipeline.Evaluate(nil)

					// Send Transaction
					d.Transactions <- transaction.InputTransaction{
						Transaction: transaction.Transaction{
							Payload: transaction.Payload{
								Reader:     reader,
								ImageBytes: nil,
							},
							ImageData:    transaction.ImageData{"count": i},
							ResponseChan: responseChan,
						},
						PipelineTag: pipeline.String(),
					}

					// Wait Transaction
					response := <-responseChan

					d.logger.Debugw("RECEIVED RESPONSE.", "ack", response.Ack, "error", response.Error)
				}(d.metric)

				d.metric++
				time.Sleep(time.Millisecond * 500)
			}
		}
	}()

	return nil
}

// Close closes the plugin gracefully
func (d *Dummy) Close(time.Duration) error {
	d.logger.Debugw("Sending closing signal...")
	d.stopChan <- struct{}{}
	d.wg.Wait()
	close(d.Transactions)
	d.logger.Debugw("Closed.")
	return nil
}
