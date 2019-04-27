package node

import (
	"context"

	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

// TODO enhance interfaces, avoid buffering reader if it's only one node in next.

//Node A node wraps components and manage receiving transactions and forwarding transactions to next nodes.
type Node struct {
	Component    Jobable
	RecieverChan chan transaction.Transaction
	Next         []NextNode
	Resource     resource.Resource
}

//Jobable Each Component wrappers should implement this interface according to how it should process a transaction.
type Jobable interface {
	Job(t transaction.Transaction) (mirror.ReaderCloner, transaction.ImageBytes, response.Response)
}

//NextNode Wraps the next node plus its attributes.
type NextNode struct {
	Async bool
	Node  Node
}

//Start Start accepting Jobs
func (n *Node) Start() {
	go func() {
		for value := range n.RecieverChan {
			go n.Job(value)
		}
	}()
}

//Job Manages processing transaction by acquiring/releasing resources, calling components job() method, and forward
//results to next nodes accordingly, and waits for response from all next Nodes, finally returning response of the transaction.
func (n *Node) Job(t transaction.Transaction) {

	////////////////////////////////////////////
	// Acquire Resource (limit concurrency)
	err := n.acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	////////////////////////////////////////////
	// Do Main Component Job
	readerCloner, ImageBytes, Response := n.Component.Job(t)
	n.release()

	if !Response.Ack {
		t.ResponseChan <- Response
		return
	}

	////////////////////////////////////////////
	// Send result to next nodes
	responseChan := make(chan response.Response, len(n.Next))
	ctx, cancel := context.WithCancel(t.Context)

	for _, next := range n.Next {
		next.Node.RecieverChan <- transaction.Transaction{
			Payload: transaction.Payload{
				Reader:     readerCloner.NewReader(),
				ImageBytes: ImageBytes,
			},
			ImageData:    t.ImageData,
			Context:      ctx,
			ResponseChan: responseChan,
		}
	}

	////////////////////////////////////////////
	// AWAIT RESPONSEEs
	count, total := 0, len(n.Next)
loop:
	for ; count < total; count++ {
		select {
		case Response = <-responseChan:
			if !Response.Ack {
				// cancel sub-contexts
				cancel()
				break loop
			}
		case <-t.Context.Done():
			// cancel sub-contexts
			cancel()
			Response = response.NoAck(t.Context.Err())
			break loop
		}
	}

	// cancel context (avoid context leak)
	cancel()

	// Send Response back.
	t.ResponseChan <- Response
}

//GetReceiverChan Return Nodes receiver channel
func (n *Node) GetReceiverChan() chan transaction.Transaction {
	return n.RecieverChan
}

func (n *Node) acquire(c context.Context) error {
	select {
	case <-c.Done():
		return c.Err()
	default:
		acquired := n.Resource.TryAcquire(1)
		if !acquired {
			// Warn for filled Node.
			n.Resource.Logger.Warn("plugin reached its concurrency limit")

			// block until acquired
			return n.Resource.Acquire(c, 1)
		}
	}

	return nil
}

func (n *Node) release() {
	n.Resource.Release(1)
}

//func (n *Node) sendToNextNodes(rb mirror.ReaderCloner, ctx context.Context, ib transaction.ImageBytes,
//	id transaction.ImageData, rc chan response.Response) {

//}

//func (n *Node) receiveResponseFromNextNodes(ResponseChan chan response.Response) response.Response {

//}
