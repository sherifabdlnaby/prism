package node

import (
	"context"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//ReadWrite Wraps a readwrite component
type ReadWrite struct {
	component.ProcessorReadWrite
	receiveChan <-chan transaction.Transaction
	Next        []Next
	Resource    resource.Resource
}

//startMux startMux receiving transactions
func (n *ReadWrite) Start() error {
	go func() {
		for value := range n.receiveChan {
			go n.job(value)
		}
	}()
	return nil
}

func (n *ReadWrite) Stop() error {
	for _, value := range n.Next {
		close(value.TransactionChan)
	}
	return nil
}

//job Process transaction by calling Decode-> Process-> Encode->
func (n *ReadWrite) job(t transaction.Transaction) {

	////////////////////////////////////////////
	// Acquire Resource (limit concurrency)
	err := n.Resource.Acquire(t.Context, 1)
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
		n.Resource.Release(1)
		return
	}

	/// PROCESS
	decodedPayload, Response := n.Process(decoded, t.ImageData)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.Resource.Release(1)
		return
	}

	var baseOutput = transaction.OutputPayload{}
	responseChan := make(chan response.Response, len(n.Next))
	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	// base Output writerCloner
	buffer := buffersPool.Get()
	defer buffersPool.Put(buffer)
	writerCloner := mirror.NewWriter(buffer)

	baseOutput = transaction.OutputPayload{
		WriteCloser: writerCloner,
		ImageBytes:  nil,
	}

	/// ENCODE
	Response = n.Encode(decodedPayload, t.ImageData, &baseOutput)
	n.Resource.Release(1)
	if !Response.Ack {
		t.ResponseChan <- Response
		return
	}

	for _, next := range n.Next {
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
	count, total := 0, len(n.Next)

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

func (n *ReadWrite) SetTransactionChan(tc <-chan transaction.Transaction) {
	n.receiveChan = tc
}
