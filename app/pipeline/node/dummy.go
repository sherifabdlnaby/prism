package node

import (
	"context"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/bufferspool"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//Dummy Used at the start of every pipeline.
type Dummy struct {
	receiveChan <-chan transaction.Transaction
	nexts       []Next
	Resource    resource.Resource
}

//startMux startMux receiving transactions
func (n *Dummy) Start() error {
	go func() {
		for value := range n.receiveChan {
			go n.job(value)
		}
	}()
	return nil
}

func (n *Dummy) Stop() error {

	for _, value := range n.nexts {
		// close this next-node chan
		close(value.TransactionChan)

		// tell this next-node to stop which in turn will close all its next(s) too.
		err := value.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

//job Just forwards the input.
func (n *Dummy) job(t transaction.Transaction) {

	////////////////////////////////////////////
	// Acquire Resource (limit concurrency)
	err := n.Resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	////////////////////////////////////////////
	// DUMMY NODE WON'T DO WORK SO JUST FORWARD.

	////////////////////////////////////////////
	// Send to next nodes
	responseChan := make(chan response.Response, len(n.nexts))
	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	if len(n.nexts) == 1 {
		n.nexts[0].TransactionChan <- transaction.Transaction{
			Payload: transaction.Payload{
				Reader:     t.Reader,
				ImageBytes: t.ImageBytes,
			},
			ImageData:    t.ImageData,
			Context:      ctx,
			ResponseChan: responseChan,
		}
	} else {
		// Create a reader cloner
		buffer := bufferspool.Get()
		defer bufferspool.Put(buffer)
		readerCloner := mirror.NewReader(t.Reader, buffer)

		for _, next := range n.nexts {
			next.TransactionChan <- transaction.Transaction{
				Payload: transaction.Payload{
					Reader:     readerCloner.Clone(),
					ImageBytes: t.ImageBytes,
				},
				ImageData:    t.ImageData,
				Context:      ctx,
				ResponseChan: responseChan,
			}
		}
	}

	////////////////////////////////////////////
	// receive from next nodes
	count, total := 0, len(n.nexts)
	var Response response.Response

loop:
	for ; count < total; count++ {
		select {
		case Response = <-responseChan:
			if !Response.Ack {
				break loop
			}
		case <-t.Context.Done():
			Response = response.NoAck(t.Context.Err())
			break loop
		}
	}

	// Send Response back.
	t.ResponseChan <- Response

	// Dummy Node release after receive response as it is used to limit the entire pipeline concurrency.
	n.Resource.Release()
}

func (n *Dummy) SetTransactionChan(tc <-chan transaction.Transaction) {
	n.receiveChan = tc
}

func (n *Dummy) SetNexts(nexts []Next) {
	n.nexts = nexts
}

func (n *Dummy) SetAsync(async bool) {
	// Dummy can't be async.
	return
}
