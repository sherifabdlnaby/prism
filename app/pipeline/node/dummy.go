package node

import (
	"context"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/bufferspool"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//dummy Used at the start of every pipeline.
type dummy struct {
	*base
}

//NewDummy Construct a new Dummy Node
func NewDummy(r resource.Resource) Node {
	Node := &dummy{}
	base := newBase(Node, r)
	Node.base = base
	return Node
}

func (n *dummy) getInternalType() interface{} {
	return nil
}

//job Just forwards the input.
func (n *dummy) job(t transaction.Transaction) {

	////////////////////////////////////////////
	// Acquire resource (limit concurrency of entire pipeline)
	err := n.resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	////////////////////////////////////////////
	// DUMMY NODE WON'T DO WORK SO JUST FORWARD.

	////////////////////////////////////////////
	// Send to next nodes

	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	responseChan := n.sendNexts(ctx, t.Payload.(payload.Bytes), t.Data, nil)

	// Await Responses
	Response := n.waitResponses(ctx, responseChan)

	// Send Response back.
	t.ResponseChan <- Response

	// dummy Node release after receive response as it is used to limit the entire pipeline concurrency.
	n.resource.Release()
}

func (n *dummy) jobStream(t transaction.Transaction) {

	////////////////////////////////////////////
	// Acquire resource (limit concurrency of entire pipeline)
	err := n.resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	////////////////////////////////////////////
	// DUMMY NODE WON'T DO WORK SO JUST FORWARD.

	////////////////////////////////////////////
	// Send to next nodes
	var responseChan chan response.Response

	// Micro Optimization
	if len(n.nexts) == 1 {
		// micro optimization. no need to put buffer cloner in-front of a single node
		responseChan = make(chan response.Response)
		n.nexts[0].TransactionChan <- transaction.Transaction{
			Payload:      t.Payload,
			Data:         t.Data,
			Context:      t.Context,
			ResponseChan: responseChan,
		}
	} else {
		// Get Buffer from pool
		buffer := bufferspool.Get()
		defer bufferspool.Put(buffer)

		// Create a reader cloner from incoming stream (to clone the reader stream as it comes in)
		stream := t.Payload.(payload.Stream)
		readerCloner := mirror.NewReader(stream, buffer)

		responseChan = n.sendNextsStream(t.Context, readerCloner, t.Data, nil)
	}

	// Await Responses
	Response := n.waitResponses(t.Context, responseChan)

	// Send Response back.
	t.ResponseChan <- Response

	// dummy Node release after receive response as it is used to limit the entire pipeline concurrency.
	n.resource.Release()
}
