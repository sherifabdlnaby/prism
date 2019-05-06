package node

import (
	"context"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//ReadOnly Wraps a ReadOnly component
type ReadOnly struct {
	component.ProcessorReadOnly
	ReceiverChan chan transaction.Transaction
	Next         []Next
	Resource     resource.Resource
}

func (n *ReadOnly) Start() {
	go func() {
		for value := range n.ReceiverChan {
			go n.job(value)
		}
	}()
}

func (n *ReadOnly) GetReceiverChan() chan transaction.Transaction {
	return n.ReceiverChan
}

//job Process transaction by calling Decode-> Process-> Encode->
func (n *ReadOnly) job(t transaction.Transaction) {

	////////////////////////////////////////////
	// Acquire Resource (limit concurrency)
	err := n.Resource.Acquire(t.Context, 1)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		n.Resource.Release(1)
		return
	}

	////////////////////////////////////////////
	// PROCESS ( DECODE -> PROCESS )

	readerCloner := mirror.NewReader(t.Payload.Reader)
	mirrorPayload := transaction.Payload{
		Reader:     readerCloner.Clone(),
		ImageBytes: t.ImageBytes,
	}

	/// DECODE
	decoded, Response := n.Decode(mirrorPayload, t.ImageData)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.Resource.Release(1)
		return
	}

	/// PROCESS
	Response = n.Process(decoded, t.ImageData)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.Resource.Release(1)
		return
	}

	n.Resource.Release(1)
	responseChan := make(chan response.Response, len(n.Next))
	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	////////////////////////////////////////////
	// forward to next nodes
	for _, next := range n.Next {
		next.GetReceiverChan() <- transaction.Transaction{
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
