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

//job Process transaction by calling Decode-> Process-> Encode->
func (n *readWrite) job(t transaction.Transaction) {

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

	var baseOutput = payload.Output{}
	responseChan := make(chan response.Response, len(n.nexts))
	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	// base output writerCloner
	buffer := bufferspool.Get()
	defer bufferspool.Put(buffer)
	writerCloner := mirror.NewWriter(buffer)

	baseOutput = payload.Output{
		WriteCloser: writerCloner,
		Bytes:       nil,
	}

	/// ENCODE
	Response = n.processor.Encode(decodedPayload, t.Data, &baseOutput)
	n.resource.Release()
	if !Response.Ack {
		t.ResponseChan <- Response
		return
	}

	for _, next := range n.nexts {
		next.TransactionChan <- transaction.Transaction{
			Payload: payload.Payload{
				Reader: writerCloner.Clone(),
				Bytes:  baseOutput.Bytes,
			},
			Data:         t.Data,
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
