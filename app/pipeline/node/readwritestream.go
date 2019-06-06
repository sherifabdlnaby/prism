package node

import (
	"context"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/bufferspool"
	"github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//readWrite Wraps a readwrite component
type readWriteStream struct {
	processor processor.ReadWriteStream
	*base
}

//NewReadWrite Construct a new ReadWrite Node
func NewReadWriteStream(processorReadWrite processor.ReadWriteStream, r resource.Resource) Node {
	Node := &readWriteStream{processor: processorReadWrite}
	base := newBase(Node, r)
	Node.base = base
	return Node
}

//job Process transaction by calling Decode-> Process-> Encode->
func (n *readWriteStream) job(t transaction.Transaction) {

	////////////////////////////////////////////
	// Acquire resource (limit concurrency)
	err := n.resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	////////////////////////////////////////////
	// PROCESS ( DECODE -> PROCESS -> ENCODE )

	/// DECODE
	decoded, Response := n.processor.Decode(t.Payload, t.Data)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.resource.Release()
		return
	}

	/// PROCESS
	decodedPayload, Response := n.processor.Process(decoded, t.Data)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.resource.Release()
		return
	}

	// base output writerCloner
	buffer := bufferspool.Get()
	defer bufferspool.Put(buffer)
	writerCloner := mirror.NewWriter(buffer)

	/// ENCODE
	Response = n.processor.EncodeStream(decodedPayload, t.Data, writerCloner)
	n.resource.Release()
	if !Response.Ack {
		t.ResponseChan <- Response
		return
	}

	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	// Send to nexts
	responseChan := n.sendNextsStream(writerCloner, t.Data, ctx)

	// Await Responses
	Response = n.waitResponses(responseChan, ctx)

	// Send Response back.
	t.ResponseChan <- Response
}

func (n *readWriteStream) jobStream(t transaction.Streamable) {

	////////////////////////////////////////////
	// Acquire resource (limit concurrency)
	err := n.resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	////////////////////////////////////////////
	// PROCESS ( DECODE -> PROCESS -> ENCODE )

	/// DECODE
	decoded, Response := n.processor.DecodeStream(t.Payload, t.Data)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.resource.Release()
		return
	}

	/// PROCESS
	decodedPayload, Response := n.processor.Process(decoded, t.Data)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.resource.Release()
		return
	}

	// base output writerCloner
	buffer := bufferspool.Get()
	defer bufferspool.Put(buffer)
	writerCloner := mirror.NewWriter(buffer)

	/// ENCODE
	Response = n.processor.EncodeStream(decodedPayload, t.Data, writerCloner)
	n.resource.Release()
	if !Response.Ack {
		t.ResponseChan <- Response
		return
	}

	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	// Send to nexts
	responseChan := n.sendNextsStream(writerCloner, t.Data, ctx)

	// Await Responses
	Response = n.waitResponses(responseChan, ctx)

	// Send Response back.
	t.ResponseChan <- Response
}
