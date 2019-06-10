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
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

// Dummy Input that read a file from root just for testing.
type Dummy struct {
	config       Config
	Transactions chan transaction.InputTransaction
	stopChan     chan struct{}
	logger       zap.SugaredLogger
	wg           sync.WaitGroup
	metric       int
}

type Config struct {
	FileName string
	Pipeline string
	Tick     int
	Timeout  int

	pipeline config.Selector
	filename config.Selector
}

func DefaultConfig() *Config {
	return &Config{
		Tick: 1000,
	}
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &Dummy{}
}

// InputTransactionChan Return Transaction Chan used to send transaction to this Component
func (d *Dummy) InputTransactionChan() <-chan transaction.InputTransaction {
	return d.Transactions
}

// Init Initializes Plugin
func (d *Dummy) Init(config config.Config, logger zap.SugaredLogger) error {
	var err error

	d.config = *DefaultConfig()
	err = config.Populate(&d.config)
	if err != nil {
		return err
	}

	d.config.filename, err = config.NewSelector(d.config.FileName)
	if err != nil {
		return err
	}

	d.config.pipeline, err = config.NewSelector(d.config.Pipeline)
	if err != nil {
		return err
	}

	d.Transactions = make(chan transaction.InputTransaction)
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

					// create context
					ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.config.Timeout)*time.Millisecond)
					defer cancel()

					// payloadData (request params)
					payloadData := payload.Data{
						"count": i,
					}

					// get selectors
					pipeline, err := d.config.pipeline.Evaluate(payloadData)
					if err != nil {
						d.logger.Debugw("Error in dummy: ", zap.Error(err))
						return
					}

					filename, err := d.config.filename.Evaluate(payloadData)
					if err != nil {
						d.logger.Debugw("Error in dummy: ", zap.Error(err))
						return
					}

					// Get Image Data
					reader, err := os.Open(filename)
					if err != nil {
						d.logger.Debugw("Error in dummy: ", zap.Error(err))
						return
					}

					// Send
					responseChan := make(chan response.Response)
					if flag {
						bytes, _ := ioutil.ReadAll(reader)
						// Send Transaction
						d.logger.Debugw("SENDING A TRANSACTION (bytes)....", "ID", i)
						d.Transactions <- transaction.InputTransaction{
							Transaction: transaction.Transaction{
								Payload:      payload.Bytes(bytes),
								Data:         payloadData,
								ResponseChan: responseChan,
								Context:      ctx,
							},
							PipelineTag: pipeline,
						}
					} else {
						d.logger.Debugw("SENDING A TRANSACTION (stream)...", "ID", i)
						d.Transactions <- transaction.InputTransaction{
							Transaction: transaction.Transaction{
								Payload:      reader,
								Data:         payloadData,
								ResponseChan: responseChan,
								Context:      ctx,
							},
							PipelineTag: pipeline,
						}
					}

					// alternate between stream/data
					flag = !flag

					// Wait Transaction
					response := <-responseChan

					d.logger.Debugw("RECEIVED RESPONSE.", "ID", i, "ack", response.Ack, "error", response.Error, "AckErr", response.AckErr)
				}(d.metric)

				d.metric++
				time.Sleep(time.Millisecond * time.Duration(d.config.Tick))
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
	d.logger.Debugw("closed.")
	return nil
}
