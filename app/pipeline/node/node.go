package node

import (
	"context"
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//Node A node wraps components and manage receiving transactions and forwarding transactions to next nodes.
type Node struct {
	Component    Jobable
	RecieverChan chan transaction.Transaction
	Next         []NextNode
	Resource     resource.Resource
}

//Jobable Each Component wrappers should implement this interface according to how it should process a transaction.
type Jobable interface {
	Job(t transaction.Transaction) (mirror.ReaderCloner, transaction.ImageBytes, transaction.Response)
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

//Job Manages processing transaction by acquiring/releasing resources, calling components Job() method, and forward
//results to next nodes accordingly, and waits for response from all next Nodes, finally returning response of the transaction.
func (n *Node) Job(t transaction.Transaction) {
	err := n.acquire(context.TODO())
	if err != nil {
		t.ResponseChan <- transaction.ResponseError(err)
		n.release()
		return
	}

	readerCloner, ImageBytes, response := n.Component.Job(t)

	if !response.Ack {
		t.ResponseChan <- response
		n.release()
		return
	}
	n.release()

	// SEND TO NEXT
	responseChan := make(chan transaction.Response)
	n.sendToNextNodes(readerCloner, ImageBytes, t.ImageData, responseChan)

	// AWAIT RESPONSEEs
	response = n.receiveResponseFromNextNodes(response, responseChan)

	// Send Response back.
	t.ResponseChan <- response
}

//GetReceiverChan Return Nodes receiver channel
func (n *Node) GetReceiverChan() chan transaction.Transaction {
	return n.RecieverChan
}

func (n *Node) sendToNextNodes(readerBase mirror.ReaderCloner, ImageBytes transaction.ImageBytes, imageData transaction.ImageData, responseChan chan transaction.Response) {
	for _, next := range n.Next {
		next.Node.RecieverChan <- transaction.Transaction{
			Payload: transaction.Payload{
				Reader:     readerBase.NewReader(),
				ImageBytes: ImageBytes,
			},
			ImageData:    imageData,
			ResponseChan: responseChan,
		}
	}
}

func (n *Node) receiveResponseFromNextNodes(response transaction.Response, responseChan chan transaction.Response) transaction.Response {
	response = transaction.ResponseACK
	count, total := 0, len(n.Next)
forloop:
	for ; count < total; count++ {
		select {
		case response = <-responseChan:
			if !response.Ack {
				break forloop
			}
			// TODO case context canceled.
		}
	}
	return response
}

func (n *Node) acquire(c context.Context) error {
	acquired := n.Resource.TryAcquire(1)
	if !acquired {
		// Warn for filled Node.
		n.Resource.Logger.Warn("plugin reached its concurrency limit")
		return n.Resource.Acquire(c, 1)
	}
	return nil
}

func (n *Node) release() {
	n.Resource.Release(1)
}
