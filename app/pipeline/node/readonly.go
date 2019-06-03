package node

import (
	"context"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/bufferspool"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//readOnly Wraps a readOnly component
type readOnly struct {
	processor component.ProcessorReadOnly
	*base
}

//NewReadOnly Construct a new ReadOnly node
func NewReadOnly(ProcessorReadOnly component.ProcessorReadOnly, resource resource.Resource) Node {
	Node := &readOnly{processor: ProcessorReadOnly}
	base := newBase(Node, resource)
	Node.base = base
	return Node
}

//job Process transaction by calling Decode-> Process-> Encode->
func (n *readOnly) job(t transaction.Transaction) {

	////////////////////////////////////////////
	// Acquire resource (limit concurrency)
	err := n.resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
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
	decoded, Response := n.processor.Decode(mirrorPayload, t.ImageData)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.resource.Release()
		return
	}

	/// PROCESS
	Response = n.processor.Process(decoded, t.ImageData)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.resource.Release()
		return
	}

	n.resource.Release()
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
