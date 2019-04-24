package node

import (
	"context"
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

type Node struct {
	Component    Jobable
	RecieverChan chan transaction.Transaction
	Next         []NextNode
	Resource     resource.Resource
}

type Jobable interface {
	Job(t transaction.Transaction) (mirror.ReaderCloner, transaction.ImageBytes, transaction.Response)
}

type NextNode struct {
	Async bool
	Node  Node
}

func (n *Node) Start() {
	go func() {
		for value := range n.RecieverChan {
			go n.Job(value)
		}
	}()
}

func (n *Node) Job(t transaction.Transaction) {
	err := n.Acquire(context.TODO())
	if err != nil {
		t.ResponseChan <- transaction.ResponseError(err)
		n.Release()
		return
	}

	readerCloner, ImageBytes, response := n.Component.Job(t)

	if !response.Ack {
		t.ResponseChan <- response
		n.Release()
		return
	}
	n.Release()

	// SEND TO NEXT
	responseChan := make(chan transaction.Response)
	n.SendToNextNodes(readerCloner, ImageBytes, t.ImageData, responseChan)

	// AWAIT RESPONSEEs
	response = n.ReceiveResponseFromNextNodes(response, responseChan)

	// Send Response back.
	t.ResponseChan <- response
}

// TODO refactor this to be better. and get rid of this virtual one. srry future me.
func (n *Node) GetReceiverChan() chan transaction.Transaction {
	return n.RecieverChan
}

func (n *Node) SendToNextNodes(readerBase mirror.ReaderCloner, ImageBytes transaction.ImageBytes, imageData transaction.ImageData, responseChan chan transaction.Response) {
	for _, next := range n.Next {
		next.Node.GetReceiverChan() <- transaction.Transaction{
			Payload: transaction.Payload{
				Reader:     readerBase.NewReader(),
				ImageBytes: ImageBytes,
			},
			ImageData:    imageData,
			ResponseChan: responseChan,
		}
	}
}

func (n *Node) ReceiveResponseFromNextNodes(response transaction.Response, responseChan chan transaction.Response) transaction.Response {
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

func (n *Node) Acquire(c context.Context) error {
	acquired := n.Resource.TryAcquire(1)
	if !acquired {
		// Warn for filled Node.
		n.Resource.Logger.Warn("plugin reached its concurrency limit")
		return n.Resource.Acquire(c, 1)
	}
	return nil
}

func (n *Node) Release() {
	n.Resource.Release(1)
}
