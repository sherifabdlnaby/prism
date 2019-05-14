package node

import (
	"context"
	"sync"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/bufferspool"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//ReadWrite Wraps a readwrite component
type ReadWrite struct {
	component.ProcessorReadWrite
	receiveChan    <-chan transaction.Transaction
	asyncResponses chan response.Response
	async          bool
	wg             sync.WaitGroup
	nexts          []Next
	Resource       resource.Resource
}

// Start starts this node and all its next nodes to start receiving transactions
// By starting all next nodes, start async request handler, and start receiving transactions
func (n *ReadWrite) Start() error {
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
func (n *ReadWrite) Stop() error {
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

//job Process transaction by calling Decode-> Process-> Encode->
func (n *ReadWrite) job(t transaction.Transaction) {

	////////////////////////////////////////////
	// Acquire Resource (limit concurrency)
	err := n.Resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	////////////////////////////////////////////
	// PROCESS ( DECODE -> PROCESS -> ENCODE )

	/// DECODE
	decoded, Response := n.Decode(t.Payload, t.ImageData)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.Resource.Release()
		return
	}

	/// PROCESS
	decodedPayload, Response := n.Process(decoded, t.ImageData)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.Resource.Release()
		return
	}

	var baseOutput = transaction.OutputPayload{}
	responseChan := make(chan response.Response, len(n.nexts))
	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	// base Output writerCloner
	buffer := bufferspool.Get()
	defer bufferspool.Put(buffer)
	writerCloner := mirror.NewWriter(buffer)

	baseOutput = transaction.OutputPayload{
		WriteCloser: writerCloner,
		ImageBytes:  nil,
	}

	/// ENCODE
	Response = n.Encode(decodedPayload, t.ImageData, &baseOutput)
	n.Resource.Release()
	if !Response.Ack {
		t.ResponseChan <- Response
		return
	}

	for _, next := range n.nexts {
		next.TransactionChan <- transaction.Transaction{
			Payload: transaction.Payload{
				Reader:     writerCloner.Clone(),
				ImageBytes: baseOutput.ImageBytes,
			},
			ImageData:    t.ImageData,
			Context:      ctx,
			ResponseChan: responseChan,
		}
	}

	////////////////////////////////////////////
	// receive from next nodes
	count, total := 0, len(n.nexts)

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
}

//SetTransactionChan Set the transaction chan node will use to receive input
func (n *ReadWrite) SetTransactionChan(tc <-chan transaction.Transaction) {
	n.receiveChan = tc
}

//SetAsync Set if this node is sync/async
func (n *ReadWrite) SetAsync(async bool) {
	if async {
		n.asyncResponses = make(chan response.Response)
	}
	n.async = async
}

//SetNexts Set this node's next nodes.
func (n *ReadWrite) SetNexts(nexts []Next) {
	n.nexts = nexts
}

// TODO to handle failing/success async requests later.
func (n *ReadWrite) startAsyncHandling() {
	go func() {
		for range n.asyncResponses {
			n.wg.Done()
		}
	}()
}
