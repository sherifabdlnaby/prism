package pipeline

import (
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//TODO refactor to make output and process close to each other.

type node struct {
	RecieverChan chan transaction.Transaction
	Next         []nextNode
	resource.Resource
}

type nextNode struct {
	Async bool
	node  Interface
}

type Interface interface {
	Start()
	GetReceiverChan() chan transaction.Transaction
	Job(t transaction.Transaction)
}

///////////////

type dummyNode struct {
	node
}

///////////////

func (dn *dummyNode) Start() {
	go func() {
		for value := range dn.RecieverChan {
			go dn.Job(value)
		}
	}()
}

func (dn *dummyNode) Job(t transaction.Transaction) {

	// SEND
	responseChan := make(chan transaction.Response)

	dn.Next[0].node.GetReceiverChan() <- transaction.Transaction{
		Payload:      t.Payload,
		ImageData:    t.ImageData,
		ResponseChan: responseChan,
	}

	// Send Response back.
	t.ResponseChan <- <-responseChan
}

// TODO refactor this to be better. and get rid of this virtual one. srry future me.
func (n node) GetReceiverChan() chan transaction.Transaction {
	return n.RecieverChan
}

func (n node) Job(t transaction.Transaction) {
	panic("virtual- this should never be called")
}

//

func (n *node) sendToNextNodes(readerBase mirror.ReaderCloner, ImageBytes transaction.ImageBytes, imageData transaction.ImageData, responseChan chan transaction.Response) {
	for _, next := range n.Next {
		next.node.GetReceiverChan() <- transaction.Transaction{
			Payload: transaction.Payload{
				Reader:     readerBase.NewReader(),
				ImageBytes: ImageBytes,
			},
			ImageData:    imageData,
			ResponseChan: responseChan,
		}
	}
}

//

func (n *node) receiveResponseFromNextNodes(response transaction.Response, responseChan chan transaction.Response) transaction.Response {
	response = transaction.ResponseACK
	count, total := 0, len(n.Next)
	for ; count < total; count++ {
		select {
		case response = <-responseChan:
			if !response.Ack {
				break
			}
			// TODO case context canceled.
		}
	}
	return response
}
