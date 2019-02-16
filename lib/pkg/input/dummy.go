package input

import (
	"github.com/sherifabdlnaby/prism/lib/pkg/types"
	"log"
	"os"
	"time"
)

// Dummy Input that read a file from root just for testing.
type Dummy struct {
	FileName     string
	Transactions chan types.StreamableTransaction
	stopChan     chan struct{}
}

func (d *Dummy) TransactionChan() <-chan types.StreamableTransaction {
	return d.Transactions
}

func (d *Dummy) Init(config types.Config) error {
	d.FileName = config["filename"].(string)
	d.Transactions = make(chan types.StreamableTransaction, 1)
	log.Println("Initialized dummy input.")
	d.stopChan = make(chan struct{})
	return nil
}

func (d *Dummy) Start() error {
	log.Println("Started Input, Hooray!")

	go func() {
		for {
			select {
			case <-d.stopChan:
				log.Println("Closing...")
				break
			default:
				log.Println("SENDING TRANSACTION...")
				reader, err := os.Open(d.FileName)
				responseChan := make(chan types.Response)

				if err != nil {
					log.Println("Error in input: ", err)
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

				log.Println("RECEIVED RESPONSE.", response)

				time.Sleep(time.Second * 3)
			}
		}
	}()

	return nil
}

func (d *Dummy) Close(time.Duration) error {
	log.Println("Sending closing signal...")
	d.stopChan <- struct{}{}
	close(d.Transactions)
	log.Println("Closed.")
	return nil
}
