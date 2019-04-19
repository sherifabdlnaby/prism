package dummy

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"go.uber.org/zap"
	"os"
	"sync"
	"time"
)

// Dummy Input that read a file from root just for testing.
type Dummy struct {
	FileName     string
	Transactions chan component.Transaction
	stopChan     chan struct{}
	logger       zap.SugaredLogger
	wg           sync.WaitGroup
	metric       int
}

func NewComponent() component.Component {
	return &Dummy{}
}

func (d *Dummy) TransactionChan() <-chan component.Transaction {
	return d.Transactions
}

func (d *Dummy) Init(config component.Config, logger zap.SugaredLogger) error {
	FileName, err := config.Get("filename", nil)
	if err != nil {
		return err
	}

	d.FileName = FileName.String()

	d.Transactions = make(chan component.Transaction, 1)
	d.stopChan = make(chan struct{})
	d.logger = logger
	return nil
}

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
					reader, err := os.Open(d.FileName)
					responseChan := make(chan component.Response)

					if err != nil {
						d.logger.Debugw("Error in dummy: ", zap.Error(err))
						return
					}

					// Send Transaction
					d.Transactions <- component.Transaction{
						InputPayload: component.InputPayload{
							Reader: reader,
						},
						ImageData:    component.ImageData{"count": i},
						ResponseChan: responseChan,
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

func (d *Dummy) Close(time.Duration) error {
	d.logger.Debugw("Sending closing signal...")
	d.stopChan <- struct{}{}
	d.wg.Wait()
	close(d.Transactions)
	d.logger.Debugw("Closed.")
	return nil
}
