package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

type base struct {
	receiveTxnChan <-chan transaction.Transaction
	asyncResponses chan response.Response
	async          bool
	wg             sync.WaitGroup
	nexts          []Next
	resource       resource.Resource
	nodeType       nodeType
}

func newBase(nodeType nodeType, resource resource.Resource) *base {
	return &base{
		asyncResponses: make(chan response.Response),
		async:          false,
		wg:             sync.WaitGroup{},
		nexts:          nil,
		resource:       resource,
		nodeType:       nodeType,
	}
}

// Start starts this nodeType and all its next nodes to start receiving transactions
// By starting all next nodes, start async request handler, and start receiving transactions
func (n *base) Start() error {
	// Start next nodes
	for _, value := range n.nexts {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	// start Async Handler
	go n.asyncHandler()

	go func() {
		for t := range n.receiveTxnChan {
			n.handleTransaction(t)
		}
	}()

	return nil
}

//Stop Stop this Node and stop all its next nodes.
func (n *base) Stop() error {
	//wait async jobs to finish
	n.wg.Wait()

	for _, value := range n.nexts {
		// close this next-nodeType chan
		close(value.TransactionChan)

		// tell this next-nodeType to stop which in turn will close all its next(s) too.
		err := value.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

//SetTransactionChan Set the transaction chan nodeType will use to receive input
func (n *base) SetTransactionChan(tc <-chan transaction.Transaction) {
	n.receiveTxnChan = tc
}

//SetNexts Set this nodeType's next nodes.
func (n *base) SetNexts(nexts []Next) {
	n.nexts = nexts
}

//SetAsync Set if this nodeType is sync/async
func (n *base) SetAsync(async bool) {
	n.async = async
}

func (n *base) handleTransaction(t transaction.Transaction) {
	// if nodeType is set async, send ack response now,
	// and navigate actual response to asyncResponses which should handle async responses
	if n.async {
		// since it will be async, sync transaction context is irrelevant.
		// (we don't want sync nodes -that cancel transaction context when finishing to avoid ctx leak- to
		// cancel async nodes too )
		t.Context = context.Background()

		// return ack response
		t.ResponseChan <- response.Ack()

		// now actual response is given to asyncResponds that should handle async responds
		t.ResponseChan = n.asyncResponses

		// used so that stop() wait for async responses to finish. (to be actually handled later)
		n.wg.Add(1)
	}

	// Start Job according to transaction payload Type
	switch t.Payload.(type) {
	case payload.Bytes:
		go n.nodeType.job(t)
	case payload.Stream:
		go n.nodeType.jobStream(t)
	default:
		// This theoretically shouldn't happen
		t.ResponseChan <- response.Error(fmt.Errorf("invalid transaction payload type, must be payload.Bytes or payload.Stream"))
	}

}

// TODO to handle failing/success async requests later.
func (n *base) asyncHandler() {
	for range n.asyncResponses {
		n.wg.Done()
	}
}

func (n *base) waitResponses(ctx context.Context, responseChan chan response.Response) response.Response {
	////////////////////////////////////////////
	// receive from next nodes
	count, total := 0, len(n.nexts)
	Response := response.Response{}

loop:
	for ; count < total; count++ {
		select {
		case Response = <-responseChan:
			if !Response.Ack {
				break loop
			}
		case <-ctx.Done():
			Response = response.NoAck(ctx.Err())
			break loop
		}
	}

	return Response
}

func (n *base) sendNextsStream(ctx context.Context, writerCloner mirror.Cloner, data payload.Data) chan response.Response {
	responseChan := make(chan response.Response, len(n.nexts))
	for _, next := range n.nexts {
		next.TransactionChan <- transaction.Transaction{
			Payload:      writerCloner.Clone(),
			Data:         data,
			Context:      ctx,
			ResponseChan: responseChan,
		}
	}
	return responseChan
}

func (n *base) sendNexts(ctx context.Context, output payload.Bytes, data payload.Data) chan response.Response {
	responseChan := make(chan response.Response, len(n.nexts))
	for _, next := range n.nexts {
		next.TransactionChan <- transaction.Transaction{
			Payload:      output,
			Data:         data,
			Context:      ctx,
			ResponseChan: responseChan,
		}
	}
	return responseChan
}
