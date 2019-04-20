package pipeline

import (
	"context"
	"github.com/sherifabdlnaby/prism/app/manager"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
)

//TODO refactor to make output and process close to each other.

type node struct {
	RecieverChan chan component.Transaction
	Next         []nextNode
}

type nextNode struct {
	async bool
	node  nodeInterface
}

type nodeInterface interface {
	start()
	getReceiverChan() chan component.Transaction
	job(t component.Transaction)
}

///////////////

type dummyNode struct {
	node
}

type processingReadWriteNode struct {
	node
	manager.ResourceManager
	component.ProcessorReadWrite
}

type processingReadOnlyNode struct {
	node
	manager.ResourceManager
	component.ProcessorReadOnly
}

type outputNode struct {
	node
	manager.ResourceManager
	component.Output
}

///////////////

func (dn *dummyNode) start() {
	go func() {
		for value := range dn.RecieverChan {
			go dn.job(value)
		}
	}()
}

func (dn *dummyNode) job(t component.Transaction) {

	// SEND
	responseChan := make(chan component.Response)

	dn.Next[0].node.getReceiverChan() <- component.Transaction{
		InputPayload: t.InputPayload,
		ImageData:    t.ImageData,
		ResponseChan: responseChan,
	}

	// Send Response back.
	t.ResponseChan <- <-responseChan
}

//////////////

func (pn *processingReadWriteNode) start() {
	go func() {
		for value := range pn.RecieverChan {
			go pn.job(value)
		}
	}()
}

func (pn *processingReadWriteNode) job(t component.Transaction) {

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

		next.node.getReceiverChan() <- component.Transaction{
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

func (pn *processingReadOnlyNode) start() {
	go func() {
		for value := range pn.RecieverChan {
			go pn.job(value)
		}
	}()
}

func (pn *processingReadOnlyNode) job(t component.Transaction) {

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

		next.node.getReceiverChan() <- component.Transaction{
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

func (on *outputNode) start() {
	go func() {
		for value := range on.RecieverChan {
			go on.job(value)
		}
	}()
}

func (on *outputNode) job(t component.Transaction) {
	// TODO assumes output don't have NEXT.
	_ = on.ResourceManager.Acquire(context.TODO(), 1)
	// TODO check err here

	on.TransactionChan() <- t

	on.ResourceManager.Release(1)
}

//////////////

func (n node) getReceiverChan() chan component.Transaction {
	return n.RecieverChan
}

func (n node) job(t component.Transaction) {
	panic("virtual")
}
