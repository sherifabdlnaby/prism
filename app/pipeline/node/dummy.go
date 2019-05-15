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
	// Acquire resource (limit concurrency)
	err := n.resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	////////////////////////////////////////////
	// DUMMY NODE WON'T DO WORK SO JUST FORWARD.

	////////////////////////////////////////////
	// Send to next nodes
	responseChan := make(chan response.Response, len(n.nexts))
	ctx, cancel := context.WithCancel(t.Context)
	defer cancel()

	if len(n.nexts) == 1 {
		n.nexts[0].TransactionChan <- transaction.Transaction{
			Payload: transaction.Payload{
				Reader:     t.Reader,
				ImageBytes: t.ImageBytes,
			},
			ImageData:    t.ImageData,
			Context:      ctx,
			ResponseChan: responseChan,
		}
	} else {
		// Create a reader cloner
		buffer := bufferspool.Get()
		defer bufferspool.Put(buffer)
		readerCloner := mirror.NewReader(t.Reader, buffer)

		for _, next := range n.nexts {
			next.TransactionChan <- transaction.Transaction{
				Payload: transaction.Payload{
					Reader:     readerCloner.Clone(),
					ImageBytes: t.ImageBytes,
				},
				ImageData:    t.ImageData,
				Context:      ctx,
				ResponseChan: responseChan,
			}
		}
	}

	////////////////////////////////////////////
	// receive from next nodes
	count, total := 0, len(n.nexts)
	var Response response.Response

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

	// dummy Node release after receive response as it is used to limit the entire pipeline concurrency.
	n.resource.Release()
}
