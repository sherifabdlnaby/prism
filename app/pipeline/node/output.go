package node

import (
	"context"
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

// Start starts this node and all its next nodes to start receiving transactions
// By starting all next nodes, start async request handler, and start receiving transactions
func (n *Output) Start() error {

	// Start next nodes
	for _, value := range n.nexts {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	// start Async Handler
	n.startAsyncHandling()

	go func() {
		for t := range n.receiveChan {

			// if node is set async, send ack response now,
			// and navigate actual response to asyncResponses which should handle async responses
			if n.async {
				// since it will be async, sync transaction context is irrelevant.
				// (we don't want sync nodes -that cancel transaction context when finishing to avoid ctx leak- to
				// cancel async nodes too )
				t.Context = context.Background()

				// return ack response
				t.ResponseChan <- response.Ack()

				// now actual response is given to asyncResponds that should handle async responds
				t.ResponseChan = n.asyncResponses

				// used so that stop() wait for async responses to finish. (to be actually handled later)
				n.wg.Add(1)
			}

			// Start Job
			go n.job(t)
		}
	}()

	return nil
}

//Stop Stop this Node and stop all its next nodes.
func (n *Output) Stop() error {
	//wait async jobs to finish
	n.wg.Wait()

	// stop next nodes
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

//SetTransactionChan Set the transaction chan node will use to receive input
func (n *Output) SetTransactionChan(tc <-chan transaction.Transaction) {
	n.receiveChan = tc
}

//SetNexts Set this node's next nodes.
func (n *Output) SetNexts(nexts []Next) {
	n.nexts = nexts
}

//SetAsync Set if this node is sync/async
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
