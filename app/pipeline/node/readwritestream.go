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
	"go.uber.org/zap"
)

//readWrite Wraps a readwrite component
type readWriteStream struct {
	processor processor.ReadWriteStream
	*Node
}

//NewReadWriteStream Construct a new ReadWriteStream Node
func NewReadWriteStream(name string, processorReadWrite processor.ReadWriteStream, r resource.Resource, logger zap.SugaredLogger) *Node {
	Node := &readWriteStream{processor: processorReadWrite}
	base := newBase(Node, r)

	// Set attributes
	base.Name = name
	base.Logger = logger

	Node.Node = base
	return Node.Node
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
	decoded, Response := n.processor.Decode(t.Payload.(payload.Bytes), t.Data)
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

	// Node output writerCloner
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
	responseChan := n.sendNextsStream(ctx, writerCloner, t.Data)

	// Await Responses
	Response = n.waitResponses(responseChan)

	// Send Response back.
	t.ResponseChan <- Response
}

func (n *readWriteStream) jobStream(t transaction.Transaction) {

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
	stream := t.Payload.(payload.Stream)
	decoded, Response := n.processor.DecodeStream(stream, t.Data)
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

	// Node output writerCloner
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
	responseChan := n.sendNextsStream(ctx, writerCloner, t.Data)

	// Await Responses
	Response = n.waitResponses(responseChan)

	// Send Response back.
	t.ResponseChan <- Response
}
