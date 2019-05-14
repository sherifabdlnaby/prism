package node

import (
	"sync"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//Output Wraps an output component
type Output struct {
	component.Output
	receiveChan    <-chan transaction.Transaction
	asyncResponses chan response.Response
	async          bool
	wg             sync.WaitGroup
	nexts          []Next
	Resource       resource.Resource
}

//startMux startMux receiving transactions
func (n *Output) Start() error {

	// start Async Handler
	n.startAsyncHandling()

	go func() {
		for value := range n.receiveChan {
			go n.job(value)
		}
	}()
	return nil
}

func (n *Output) Stop() error {
	//wait async jobs to finish
	n.wg.Wait()

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

//job Output job will send the transaction to output plugin and await its result.
func (n *Output) job(t transaction.Transaction) {
	err := n.Resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	// if node is set async, send response now, and navigate actual response to asyncResponses which should handle async responses
	if n.async {
		t.ResponseChan <- response.Ack()

		// used so that stop() wait for async responses to finish. (may be improved later)
		n.wg.Add(1)

		t.ResponseChan = n.asyncResponses
	}

	responseChan := make(chan response.Response)

	n.TransactionChan() <- transaction.Transaction{
		Payload:      t.Payload,
		ImageData:    t.ImageData,
		ResponseChan: responseChan,
		Context:      t.Context,
	}

	t.ResponseChan <- <-responseChan

	n.Resource.Release()
}

func (n *Output) SetTransactionChan(tc <-chan transaction.Transaction) {
	n.receiveChan = tc
}

func (n *Output) SetNexts(nexts []Next) {
	n.nexts = nexts
}

func (n *Output) SetAsync(async bool) {
	if async {
		n.asyncResponses = make(chan response.Response)
	}
	n.async = async
}

// TODO to handle failing/success async requests later.
func (n *Output) startAsyncHandling() {
	go func() {
		for range n.asyncResponses {
			n.wg.Done()
		}
	}()
}
