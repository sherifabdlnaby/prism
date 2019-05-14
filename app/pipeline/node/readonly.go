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

//ReadOnly Wraps a ReadOnly component
type ReadOnly struct {
	component.ProcessorReadOnly
	receiveChan    <-chan transaction.Transaction
	asyncResponses chan response.Response
	async          bool
	wg             sync.WaitGroup
	nexts          []Next
	Resource       resource.Resource
}

//startMux startMux receiving transactions
func (n *ReadOnly) Start() error {

	// start Async Handler
	n.startAsyncHandling()

	go func() {
		for value := range n.receiveChan {
			go n.job(value)
		}
	}()

	return nil
}

func (n *ReadOnly) Stop() error {
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
func (n *ReadOnly) job(t transaction.Transaction) {

	////////////////////////////////////////////
	// Acquire Resource (limit concurrency)
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

	////////////////////////////////////////////
	// PROCESS ( DECODE -> PROCESS )

	buffer := bufferspool.Get()
	defer bufferspool.Put(buffer)
	readerCloner := mirror.NewReader(t.Payload.Reader, buffer)
	mirrorPayload := transaction.Payload{
		Reader:     readerCloner.Clone(),
		ImageBytes: t.ImageBytes,
	}

	/// DECODE
	decoded, Response := n.Decode(mirrorPayload, t.ImageData)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.Resource.Release()
		return
	}

	/// PROCESS
	Response = n.Process(decoded, t.ImageData)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.Resource.Release()
		return
	}

	n.Resource.Release()
	responseChan := make(chan response.Response, len(n.nexts))
	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	////////////////////////////////////////////
	// forward to next nodes
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

func (n *ReadOnly) SetTransactionChan(tc <-chan transaction.Transaction) {
	n.receiveChan = tc
}

func (n *ReadOnly) SetNexts(nexts []Next) {
	n.nexts = nexts
}

func (n *ReadOnly) SetAsync(async bool) {
	if async {
		n.asyncResponses = make(chan response.Response)
	}
	n.async = async
}

// TODO to handle failing/success async requests later.
func (n *ReadOnly) startAsyncHandling() {
	go func() {
		for range n.asyncResponses {
			n.wg.Done()
		}
	}()
}
