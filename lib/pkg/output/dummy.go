package output

import (
	"github.com/sherifabdlnaby/prism/lib/pkg/types"
	"io/ioutil"
	"log"
	"time"
)

// Dummy Input that read a file from root just for testing.
type Dummy struct {
	FileName     string
	Transactions chan types.Transaction
	stopChan     chan struct{}
}

func (d *Dummy) TransactionChan() chan<- types.Transaction {
	return d.Transactions
}

func (d *Dummy) Init(config types.Config) error {
	d.FileName = config["filename"].(string)
	d.Transactions = make(chan types.Transaction, 1)
	log.Println("Initialized dummy output.")
	d.stopChan = make(chan struct{})
	return nil
}

func (d *Dummy) Start() error {
	log.Println("Started Output, Hooray!")

	go func() {
		for {
			select {
			case <-d.stopChan:
				log.Println("Closing...")
				break
			case transaction := <-d.Transactions:
				log.Println("RECEIVED OUTPUT TRANSACTION...")

				err := ioutil.WriteFile(d.FileName, transaction.ImageBytes, 0644)

				if err != nil {
					log.Println("Error in output: ", err)
					continue
				}

				log.Println("OUTPUT SUCCESSFUL, Sending Response. ")

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
	log.Println("Sending closing signal...")
	d.stopChan <- struct{}{}
	close(d.Transactions)
	log.Println("Closed.")
	return nil
}
