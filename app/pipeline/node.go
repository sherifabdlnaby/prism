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

func (n node) getReceiverChan() chan component.Transaction {
	return n.RecieverChan
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

type processingNode struct {
	node
	manager.ProcessorWrapper
}

type outputNode struct {
	node
	manager.OutputWrapper
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

func (pn *processingNode) start() {
	go func() {
		for value := range pn.RecieverChan {
			go pn.job(value)
		}
	}()
}

func (pn *processingNode) job(t component.Transaction) {

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

func (pn *outputNode) start() {
	go func() {
		for value := range pn.RecieverChan {
			go pn.job(value)
		}
	}()
}

func (pn *outputNode) job(t component.Transaction) {
	// TODO assumes output don't have NEXT.
	_ = pn.OutputWrapper.Acquire(context.TODO(), 1)
	// TODO check err here

	pn.TransactionChan() <- t

	pn.OutputWrapper.Release(1)
}
