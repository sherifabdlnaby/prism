package node

import (
	"context"
	"log"
	"time"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//readWrite Wraps a readwrite component
type readWrite struct {
	processor processor.ReadWrite
	*base
}

//NewReadWrite Construct a new ReadWrite Node
func NewReadWrite(processorReadWrite processor.ReadWrite, r resource.Resource) Node {
	Node := &readWrite{processor: processorReadWrite}
	base := newBase(Node, r)
	Node.base = base
	return Node
}

func (n *readWrite) getInternalType() interface{} {
	return n.processor
}

//job Process transaction by calling Decode-> Process-> Encode->
func (n *readWrite) job(t transaction.Transaction) {

	////////////////////////////////////////////
	// Acquire resource (limit concurrency)
	err := n.resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	now := time.Now()

	////////////////////////////////////////////
	// PROCESS ( DECODE -> PROCESS -> ENCODE )

	/// DECODE
	var decoded payload.DecodedImage
	var Response response.Response

	if !t.SameType {
		decoded, Response = n.processor.Decode(t.Payload.(payload.Bytes), t.Data)
		if !Response.Ack {
			t.ResponseChan <- Response
			n.resource.Release()
			return
		}
	} else {
		decoded = t.Payload
	}

	/// PROCESS
	decodedPayload, Response := n.processor.Process(decoded, t.Data)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.resource.Release()
		return
	}

	/// NSHOF HAN ENCODE WALA LA2
	escapeEncode := true
	for _, value := range n.nexts {
		if !value.Same {
			escapeEncode = false
			break
		}
	}

	/// ENCODE
	var output payload.Bytes
	if !escapeEncode {
		output, Response = n.processor.Encode(decodedPayload, t.Data)
		n.resource.Release()
		if !Response.Ack {
			t.ResponseChan <- Response
			return
		}
	}

	log.Println(time.Now().Sub(now))

	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	// send to next channels
	responseChan := n.sendNexts(ctx, output, t.Data, decodedPayload)

	// Await Responses
	Response = n.waitResponses(ctx, responseChan)

	// Send Response back.
	t.ResponseChan <- Response
}

func (n *readWrite) jobStream(t transaction.Transaction) {

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
	var decoded payload.DecodedImage
	var Response response.Response

	now := time.Now()

	if !t.SameType {
		decoded, Response = n.processor.DecodeStream(t.Payload.(payload.Stream), t.Data)
		if !Response.Ack {
			t.ResponseChan <- Response
			n.resource.Release()
			return
		}
	} else {
		decoded = t.Payload
	}

	/// PROCESS
	decodedPayload, Response := n.processor.Process(decoded, t.Data)
	if !Response.Ack {
		t.ResponseChan <- Response
		n.resource.Release()
		return
	}

	/// NSHOF HAN ENCODE WALA LA2
	escapeEncode := true
	for _, value := range n.nexts {
		if !value.Same {
			escapeEncode = false
			break
		}
	}

	/// ENCODE
	var output payload.Bytes
	if !escapeEncode {
		output, Response = n.processor.Encode(decodedPayload, t.Data)
		n.resource.Release()
		if !Response.Ack {
			t.ResponseChan <- Response
			return
		}
	}

	log.Println(time.Now().Sub(now))

	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	// send to next channels
	responseChan := n.sendNexts(ctx, output, t.Data, decodedPayload)

	// Await Responses
	Response = n.waitResponses(ctx, responseChan)

	// Send Response back.
	t.ResponseChan <- Response
}
