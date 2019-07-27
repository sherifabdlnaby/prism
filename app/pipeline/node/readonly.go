package node

import (
	"context"

	"github.com/sherifabdlnaby/prism/pkg/bufferspool"
	"github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
)

//readOnly Wraps a readOnly core
type readOnly struct {
	processor processor.ReadOnly
	*Node
}

//process Process process by calling Decode-> Process-> Encode->
func (n *readOnly) process(t job.Job) {

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
	decoded, Response := n.processor.Decode(t.Payload.(payload.Bytes), t.Data)
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
	responseChan := n.sendNexts(ctx, t.Payload.(payload.Bytes), t.Data)

	// Await Responses
	Response = n.waitResponses(responseChan)

	// Send Response back.
	t.ResponseChan <- Response
}

func (n *readOnly) processStream(t job.Job) {

	////////////////////////////////////////////
	// Acquire resource (limit concurrency)
	err := n.resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	// Lookup Buffer from pool
	buffer := bufferspool.Get()
	defer bufferspool.Put(buffer)

	// Create a reader cloner from incoming stream (to clone the reader stream as it comes in)
	stream := t.Payload.(payload.Stream)
	readerCloner := mirror.NewReader(stream, buffer)
	var mirrorPayload payload.Stream = readerCloner.Clone()

	////////////////////////////////////////////
	// PROCESS ( DECODE -> PROCESS )

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
	responseChan := n.sendNextsStream(ctx, readerCloner, t.Data)

	// Await Responses
	Response = n.waitResponses(responseChan)

	// Send Response back.
	t.ResponseChan <- Response
}
