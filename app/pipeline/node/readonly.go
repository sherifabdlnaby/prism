package node

import (
	"context"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/bufferspool"
	"github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//readOnly Wraps a readOnly component
type readOnly struct {
	processor processor.ReadOnly
	*base
}

//NewReadOnly Construct a new ReadOnly node
func NewReadOnly(ProcessorReadOnly processor.ReadOnly, resource resource.Resource) Node {
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

	/// DECODE
	decoded, Response := n.processor.Decode(t.Payload, t.Data)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.resource.Release()
		return
	}

	/// PROCESS
	Response = n.processor.Process(decoded, t.Data)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.resource.Release()
		return
	}

	n.resource.Release()

	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	// send to next channels
	responseChan := n.sendNexts(t.Payload, t.Data, ctx)

	// Await Responses
	Response = n.waitResponses(responseChan, ctx)

	// Send Response back.
	t.ResponseChan <- Response
}

func (n *readOnly) jobStream(t transaction.Streamable) {

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
	readerCloner := mirror.NewReader(t.Payload, buffer)
	var mirrorPayload payload.Stream = readerCloner.Clone()

	/// DECODE
	decoded, Response := n.processor.DecodeStream(mirrorPayload, t.Data)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.resource.Release()
		return
	}

	/// PROCESS
	Response = n.processor.Process(decoded, t.Data)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.resource.Release()
		return
	}

	n.resource.Release()

	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	// send to next channels
	responseChan := n.sendNextsStream(readerCloner, t.Data, ctx)

	// Await Responses
	Response = n.waitResponses(responseChan, ctx)

	// Send Response back.
	t.ResponseChan <- Response
}
