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

//process process process by calling Decode-> process-> Encode->
func (n *readOnly) process(j job.Job) {

	////////////////////////////////////////////
	// Acquire resource (limit concurrency)
	err := n.resource.Acquire(j.Context)
	if err != nil {
		j.ResponseChan <- response.NoAck(err)
		return
	}

	////////////////////////////////////////////
	// PROCESS ( DECODE -> PROCESS )

	/// DECODE
	decoded, Response := n.processor.Decode(j.Payload.(payload.Bytes), j.Data)
	if !Response.Ack {
		j.ResponseChan <- Response
		n.resource.Release()
		return
	}

	/// PROCESS
	Response = n.processor.Process(decoded, j.Data)
	if !Response.Ack {
		j.ResponseChan <- Response
		n.resource.Release()
		return
	}

	n.resource.Release()

	ctx, cancel := context.WithCancel(j.Context)
	defer cancel()

	// send to next channels
	responseChan := n.sendNexts(ctx, j.Payload.(payload.Bytes), j.Data)

	// Await Responses
	Response = n.waitResponses(responseChan)

	// Send Response back.
	j.ResponseChan <- Response
}

func (n *readOnly) processStream(j job.Job) {

	////////////////////////////////////////////
	// Acquire resource (limit concurrency)
	err := n.resource.Acquire(j.Context)
	if err != nil {
		j.ResponseChan <- response.NoAck(err)
		return
	}

	// Lookup Buffer from pool
	buffer := bufferspool.Get()
	defer bufferspool.Put(buffer)

	// Create a reader cloner from incoming stream (to clone the reader stream as it comes in)
	stream := j.Payload.(payload.Stream)
	readerCloner := mirror.NewReader(stream, buffer)
	var mirrorPayload payload.Stream = readerCloner.Clone()

	////////////////////////////////////////////
	// PROCESS ( DECODE -> PROCESS )

	/// DECODE
	decoded, Response := n.processor.DecodeStream(mirrorPayload, j.Data)
	if !Response.Ack {
		j.ResponseChan <- Response
		n.resource.Release()
		return
	}

	/// PROCESS
	Response = n.processor.Process(decoded, j.Data)
	if !Response.Ack {
		j.ResponseChan <- Response
		n.resource.Release()
		return
	}

	n.resource.Release()

	ctx, cancel := context.WithCancel(j.Context)
	defer cancel()

	// send to next channels
	responseChan := n.sendNextsStream(ctx, readerCloner, j.Data)

	// Await Responses
	Response = n.waitResponses(responseChan)

	// Send Response back.
	j.ResponseChan <- Response
}
