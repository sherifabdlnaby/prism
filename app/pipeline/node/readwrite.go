package node

import (
	"context"

	"github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
)

//readWrite Wraps a readwrite Type
type readWrite struct {
	processor processor.ReadWrite
	*Node
}

//job Process job by calling Decode-> Process-> Encode->
func (n *readWrite) job(t job.Job) {

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

	/// ENCODE
	output, Response := n.processor.Encode(decodedPayload, t.Data)
	n.resource.Release()
	if !Response.Ack {
		t.ResponseChan <- Response
		return
	}

	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	// send to next channels
	responseChan := n.sendNexts(ctx, output, t.Data)

	// Await Responses
	Response = n.waitResponses(responseChan)

	// Send Response back.
	t.ResponseChan <- Response
}

func (n *readWrite) jobStream(t job.Job) {

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

	/// ENCODE
	output, Response := n.processor.Encode(decodedPayload, t.Data)
	n.resource.Release()
	if !Response.Ack {
		t.ResponseChan <- Response
		return
	}

	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	// send to next channels
	responseChan := n.sendNexts(ctx, output, t.Data)

	// Await Responses
	Response = n.waitResponses(responseChan)

	// Send Response back.
	t.ResponseChan <- Response
}
