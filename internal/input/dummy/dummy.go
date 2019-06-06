package dummy

import (
	"context"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	response2 "github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

// Dummy Input that read a file from root just for testing.
type Dummy struct {
	FileName           config.Value
	Pipeline           config.Value
	Transactions       chan transaction.Transaction
	StreamTransactions chan transaction.Streamable
	stopChan           chan struct{}
	logger             zap.SugaredLogger
	wg                 sync.WaitGroup
	metric             int
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &Dummy{}
}

// TransactionChan Return Transaction Chan used to send transaction to this Component
func (d *Dummy) TransactionChan() <-chan transaction.Transaction {
	return d.Transactions
}

// TransactionChan Return Transaction Chan used to send transaction to this Component
func (d *Dummy) StreamTransactionChan() <-chan transaction.Streamable {
	return d.StreamTransactions
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
	d.Transactions = make(chan transaction.Transaction)
	d.StreamTransactions = make(chan transaction.Streamable)
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
		flag := false
		for {
			select {
			case <-d.stopChan:
				d.logger.Debugw("closing...")
				return
			default:
				d.wg.Add(1)
				go func(i int) {
					defer d.wg.Done()

					pipeline := d.Pipeline.Get()
					ctx := context.Background()
					//defer cancel()

					filename, _ := d.FileName.Evaluate(nil)
					responseChan := make(chan response2.Response)
					reader, err := os.Open(filename.String())
					if err != nil {
						d.logger.Debugw("Error in dummy: ", zap.Error(err))
						return
					}

					// Send Transaction
					d.logger.Debugw("SENDING A TRANSACTION...", "ID", i)
					if flag {
						bytes, _ := ioutil.ReadAll(reader)
						d.Transactions <- transaction.Transaction{
							Payload:      bytes,
							Data:         payload.Data{"_pipeline": pipeline.String(), "count": i},
							ResponseChan: responseChan,
							Context:      ctx,
						}
					} else {
						d.StreamTransactions <- transaction.Streamable{
							Payload:      reader,
							Data:         payload.Data{"_pipeline": pipeline.String(), "count": i},
							ResponseChan: responseChan,
							Context:      ctx,
						}
					}

					d.logger.Debugw("SENT", "ID", i)

					//flag = !flag

					// Wait Transaction
					response := <-responseChan

					d.logger.Debugw("RECEIVED RESPONSE.", "ID", i, "ack", response.Ack, "error", response.Error, "AckErr", response.AckErr)
				}(d.metric)

				d.metric++
				time.Sleep(time.Millisecond * 5000)
			}
		}
	}()

	return nil
}

// Close closes the plugin gracefully
func (d *Dummy) Close() error {
	d.logger.Debugw("received closing signal, closing...")
	d.stopChan <- struct{}{}
	d.wg.Wait()
	close(d.Transactions)
	close(d.StreamTransactions)
	d.logger.Debugw("closed.")
	return nil
}
