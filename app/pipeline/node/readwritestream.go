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

//readWrite Wraps a readwrite core
type readWriteStream struct {
	processor processor.ReadWriteStream
	*Node
}

//process process process by calling Decode-> process-> Encode->
func (n *readWriteStream) process(j job.Job) {

	////////////////////////////////////////////
	// Acquire resource (limit concurrency)
	err := n.resource.Acquire(j.Context)
	if err != nil {
		j.ResponseChan <- response.NoAck(err)
		return
	}

	////////////////////////////////////////////
	// PROCESS ( DECODE -> PROCESS -> ENCODE )

	/// DECODE
	decoded, Response := n.processor.Decode(j.Payload.(payload.Bytes), j.Data)
	if !Response.Ack {
		j.ResponseChan <- Response
		n.resource.Release()
		return
	}

	/// PROCESS
	decodedPayload, Response := n.processor.Process(decoded, j.Data)
	if !Response.Ack {
		j.ResponseChan <- Response
		n.resource.Release()
		return
	}

	// Node output writerCloner
	buffer := bufferspool.Get()
	defer bufferspool.Put(buffer)
	writerCloner := mirror.NewWriter(buffer)

	/// ENCODE
	Response = n.processor.EncodeStream(decodedPayload, j.Data, writerCloner)
	n.resource.Release()
	if !Response.Ack {
		j.ResponseChan <- Response
		return
	}

	ctx, cancel := context.WithCancel(j.Context)
	defer cancel()

	// Send to nexts
	responseChan := n.sendNextsStream(ctx, writerCloner, j.Data)

	// Await Responses
	Response = n.waitResponses(responseChan)

	// Send Response back.
	j.ResponseChan <- Response
}

func (n *readWriteStream) processStream(j job.Job) {

	////////////////////////////////////////////
	// Acquire resource (limit concurrency)
	err := n.resource.Acquire(j.Context)
	if err != nil {
		j.ResponseChan <- response.NoAck(err)
		return
	}

	////////////////////////////////////////////
	// PROCESS ( DECODE -> PROCESS -> ENCODE )

	/// DECODE
	stream := j.Payload.(payload.Stream)
	decoded, Response := n.processor.DecodeStream(stream, j.Data)
	if !Response.Ack {
		j.ResponseChan <- Response
		n.resource.Release()
		return
	}

	/// PROCESS
	decodedPayload, Response := n.processor.Process(decoded, j.Data)
	if !Response.Ack {
		j.ResponseChan <- Response
		n.resource.Release()
		return
	}

	// Node output writerCloner
	buffer := bufferspool.Get()
	defer bufferspool.Put(buffer)
	writerCloner := mirror.NewWriter(buffer)

	/// ENCODE
	Response = n.processor.EncodeStream(decodedPayload, j.Data, writerCloner)
	n.resource.Release()
	if !Response.Ack {
		j.ResponseChan <- Response
		return
	}

	ctx, cancel := context.WithCancel(j.Context)
	defer cancel()

	// Send to nexts
	responseChan := n.sendNextsStream(ctx, writerCloner, j.Data)

	// Await Responses
	Response = n.waitResponses(responseChan)

	// Send Response back.
	j.ResponseChan <- Response
}
