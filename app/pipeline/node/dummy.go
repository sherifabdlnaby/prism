package node

import (
	"context"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/bufferspool"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
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

	responseChan := n.sendNexts(t.Payload, t.Data, ctx)

	// Await Responses
	Response := n.waitResponses(responseChan, ctx)

	// Send Response back.
	t.ResponseChan <- Response

	// dummy Node release after receive response as it is used to limit the entire pipeline concurrency.
	n.resource.Release()
}

func (n *dummy) jobStream(t transaction.Streamable) {

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
	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	if len(n.nexts) == 1 {
		// micro optimization. no need to put buffer cloner in-front of a single node
		responseChan = make(chan response.Response)
		n.nexts[0].StreamTransactionChan <- transaction.Streamable{
			Payload:      t.Payload,
			Data:         t.Data,
			Context:      ctx,
			ResponseChan: responseChan,
		}

	} else {
		// Create a reader cloner
		buffer := bufferspool.Get()
		defer bufferspool.Put(buffer)
		readerCloner := mirror.NewReader(t.Payload, buffer)

		responseChan = n.sendNextsStream(readerCloner, t.Data, ctx)
	}

	////////////////////////////////////////////

	// Await Responses
	Response := n.waitResponses(responseChan, t.Context)

	// Send Response back.
	t.ResponseChan <- Response

	// dummy Node release after receive response as it is used to limit the entire pipeline concurrency.
	n.resource.Release()
}
