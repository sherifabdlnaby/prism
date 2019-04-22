package node

import (
	"context"
	"github.com/sherifabdlnaby/prism/app/registery"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
)

//TODO refactor to make output and process close to each other.

type Node struct {
	RecieverChan chan component.Transaction
	Next         []NextNode
}

type NextNode struct {
	Async bool
	Node  Interface
}

type Interface interface {
	Start()
	GetReceiverChan() chan component.Transaction
	Job(t component.Transaction)
}

///////////////

type DummyNode struct {
	Node
}

type ProcessingReadWriteNode struct {
	Node
	registery.ResourceManager
	component.ProcessorReadWrite
}

type ProcessingReadOnlyNode struct {
	Node
	registery.ResourceManager
	component.ProcessorReadOnly
}

type OutputNode struct {
	Node
	registery.ResourceManager
	component.Output
}

///////////////

func (dn *DummyNode) Start() {
	go func() {
		for value := range dn.RecieverChan {
			go dn.Job(value)
		}
	}()
}

func (dn *DummyNode) Job(t component.Transaction) {

	// SEND
	responseChan := make(chan component.Response)

	dn.Next[0].Node.GetReceiverChan() <- component.Transaction{
		InputPayload: t.InputPayload,
		ImageData:    t.ImageData,
		ResponseChan: responseChan,
	}

	// Send Response back.
	t.ResponseChan <- <-responseChan
}

//////////////

func (pn *ProcessingReadWriteNode) Start() {
	go func() {
		for value := range pn.RecieverChan {
			go pn.Job(value)
		}
	}()
}

func (pn *ProcessingReadWriteNode) Job(t component.Transaction) {

	err := pn.ResourceManager.Acquire(context.TODO(), 1)
	if err != nil {
		t.ResponseChan <- component.ResponseError(err)
		pn.ResourceManager.Release(1)
		return
	}

	decoded, response := pn.Decode(t.InputPayload, t.ImageData)

	if !response.Ack {
		t.ResponseChan <- response
		pn.ResourceManager.Release(1)
		return
	}

	decodedPayload, response := pn.Process(decoded, t.ImageData)

	if !response.Ack {
		t.ResponseChan <- response
		pn.ResourceManager.Release(1)
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
		pn.ResourceManager.Release(1)
		return
	}

	pn.ResourceManager.Release(1)

	// SEND
	responseChan := make(chan component.Response)
	for _, next := range pn.Next {

		next.Node.GetReceiverChan() <- component.Transaction{
			InputPayload: component.InputPayload{
				Reader:     buffer.NewReader(),
				ImageBytes: baseOutput.ImageBytes,
			},
			ImageData:    t.ImageData,
			ResponseChan: responseChan,
		}
	}

	// AWAIT RESPONSEEs
	response = component.Response{}
	count, total := 0, len(pn.Next)
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

func (pn *ProcessingReadOnlyNode) Start() {
	go func() {
		for value := range pn.RecieverChan {
			go pn.Job(value)
		}
	}()
}

func (pn *ProcessingReadOnlyNode) Job(t component.Transaction) {

	err := pn.ResourceManager.Acquire(context.TODO(), 1)
	if err != nil {
		t.ResponseChan <- component.ResponseError(err)
		pn.ResourceManager.Release(1)
		return
	}

	//create reader mirror
	mrr := mirror.NewReader(t.InputPayload.Reader)

	mirrorPayload := component.InputPayload{
		Reader:     mrr.NewReader(),
		ImageBytes: t.ImageBytes,
	}
	decoded, response := pn.Decode(mirrorPayload, t.ImageData)

	if !response.Ack {
		t.ResponseChan <- response
		pn.ResourceManager.Release(1)
		return
	}

	response = pn.Process(decoded, t.ImageData)

	if !response.Ack {
		t.ResponseChan <- response
		pn.ResourceManager.Release(1)
		return
	}

	pn.ResourceManager.Release(1)

	// SEND
	responseChan := make(chan component.Response)
	for _, next := range pn.Next {

		next.Node.GetReceiverChan() <- component.Transaction{
			InputPayload: component.InputPayload{
				Reader:     mrr.NewReader(),
				ImageBytes: t.ImageBytes,
			},
			ImageData:    t.ImageData,
			ResponseChan: responseChan,
		}
	}

	// AWAIT RESPONSEEs
	response = component.Response{}
	count, total := 0, len(pn.Next)
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

func (on *OutputNode) Start() {
	go func() {
		for value := range on.RecieverChan {
			go on.Job(value)
		}
	}()
}

func (on *OutputNode) Job(t component.Transaction) {
	// TODO assumes output don't have NEXT.
	_ = on.ResourceManager.Acquire(context.TODO(), 1)
	// TODO check err here

	on.TransactionChan() <- t

	on.ResourceManager.Release(1)
}

//////////////

func (n Node) GetReceiverChan() chan component.Transaction {
	return n.RecieverChan
}

func (n Node) Job(t component.Transaction) {
	panic("virtual")
}
