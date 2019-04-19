package pipeline

import (
	"context"
	"github.com/sherifabdlnaby/prism/app/manager"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
)

//TODO refactor to make output and process close to each other.

type Node struct {
	RecieverChan chan component.Transaction
	Next         []NextNode
}

type NextNode struct {
	async bool
	Node  NodeInterface
}

func (n Node) GetRecieverChan() chan component.Transaction {
	return n.RecieverChan
}

type NodeInterface interface {
	Start()
	GetRecieverChan() chan component.Transaction
	Job(t component.Transaction)
}

///////////////

type dummyNode struct {
	Node
}

type ProcessingNode struct {
	Node
	manager.ProcessorWrapper
}

type OutputNode struct {
	Node
	manager.OutputWrapper
}

///////////////

func (dn *dummyNode) Start() {
	go func() {
		for value := range dn.RecieverChan {
			go dn.Job(value)
		}
	}()
}

func (dn *dummyNode) Job(t component.Transaction) {

	// SEND
	responseChan := make(chan component.Response)

	dn.Next[0].Node.GetRecieverChan() <- component.Transaction{
		InputPayload: t.InputPayload,
		ImageData:    t.ImageData,
		ResponseChan: responseChan,
	}

	// Send Response back.
	t.ResponseChan <- <-responseChan
}

//////////////

func (pn *ProcessingNode) Start() {
	go func() {
		for value := range pn.RecieverChan {
			go pn.Job(value)
		}
	}()
}

func (pn *ProcessingNode) Job(t component.Transaction) {

	err := pn.ProcessorWrapper.Acquire(context.TODO(), 1)
	if err != nil {
		t.ResponseChan <- component.ResponseError(err)
		pn.ProcessorWrapper.Release(1)
		return
	}

	decoded, response := pn.Decode(t.InputPayload, t.ImageData)

	if !response.Ack {
		t.ResponseChan <- response
		pn.ProcessorWrapper.Release(1)
		return
	}

	decodedPayload, response := pn.Process(decoded, t.ImageData)

	if !response.Ack {
		t.ResponseChan <- response
		pn.ProcessorWrapper.Release(1)
		return
	}

	///BASE READER
	buffer := mirror.Writer{}

	baseOutput := component.OutputPayload{
		WriteCloser: &buffer,
		ImageBytes:  nil,
	}

	response = pn.Encode(decodedPayload, t.ImageData, &baseOutput)

	if !response.Ack {
		t.ResponseChan <- response
		pn.ProcessorWrapper.Release(1)
		return
	}

	pn.ProcessorWrapper.Release(1)

	// SEND
	responseChan := make(chan component.Response)
	for _, next := range pn.Next {

		next.Node.GetRecieverChan() <- component.Transaction{
			InputPayload: component.InputPayload{
				Reader:     buffer.NewReader(),
				ImageBytes: baseOutput.ImageBytes,
			},
			ImageData:    t.ImageData,
			ResponseChan: responseChan,
		}
	}

	count, total := 0, len(pn.Next)

	// AWAIT RESPONSEEs
	response = component.Response{}
	for ; count < total; count++ {
		select {
		case response = <-responseChan:
			if !response.Ack {
				break
			}
			// TODO case context canceled.
		}
	}

	// Send Response back.
	t.ResponseChan <- response
}

//////////////

func (pn *OutputNode) Start() {
	go func() {
		for value := range pn.RecieverChan {
			go pn.Job(value)
		}
	}()
}

func (pn *OutputNode) Job(t component.Transaction) {
	// TODO assumes output don't have NEXT.
	_ = pn.OutputWrapper.Acquire(context.TODO(), 1)
	// TODO check err here

	pn.TransactionChan() <- t

	pn.OutputWrapper.Release(1)
}
