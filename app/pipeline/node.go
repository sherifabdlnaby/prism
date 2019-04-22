package pipeline

import (
	"context"
	"github.com/sherifabdlnaby/prism/app/registery/wrapper"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//TODO refactor to make output and process close to each other.

type Node struct {
	RecieverChan chan transaction.Transaction
	Next         []NextNode
}

type NextNode struct {
	Async bool
	Node  Interface
}

type Interface interface {
	Start()
	GetReceiverChan() chan transaction.Transaction
	Job(t transaction.Transaction)
}

///////////////

type DummyNode struct {
	Node
}

type ProcessingReadWriteNode struct {
	Node
	wrapper.Resource
	component.ProcessorReadWrite
}

type ProcessingReadOnlyNode struct {
	Node
	wrapper.Resource
	component.ProcessorReadOnly
}

type OutputNode struct {
	Node
	wrapper.Resource
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

func (dn *DummyNode) Job(t transaction.Transaction) {

	// SEND
	responseChan := make(chan transaction.Response)

	dn.Next[0].Node.GetReceiverChan() <- transaction.Transaction{
		Payload:      t.Payload,
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

func (pn *ProcessingReadWriteNode) Job(t transaction.Transaction) {

	err := pn.Resource.Acquire(context.TODO(), 1)
	if err != nil {
		t.ResponseChan <- transaction.ResponseError(err)
		pn.Resource.Release(1)
		return
	}

	decoded, response := pn.Decode(t.Payload, t.ImageData)

	if !response.Ack {
		t.ResponseChan <- response
		pn.Resource.Release(1)
		return
	}

	decodedPayload, response := pn.Process(decoded, t.ImageData)

	if !response.Ack {
		t.ResponseChan <- response
		pn.Resource.Release(1)
		return
	}

	///BASE READER
	buffer := mirror.Writer{}
	baseOutput := transaction.OutputPayload{
		WriteCloser: &buffer,
		ImageBytes:  nil,
	}

	response = pn.Encode(decodedPayload, t.ImageData, &baseOutput)

	if !response.Ack {
		t.ResponseChan <- response
		pn.Resource.Release(1)
		return
	}

	pn.Resource.Release(1)

	// SEND
	responseChan := make(chan transaction.Response)
	for _, next := range pn.Next {

		next.Node.GetReceiverChan() <- transaction.Transaction{
			Payload: transaction.Payload{
				Reader:     buffer.NewReader(),
				ImageBytes: baseOutput.ImageBytes,
			},
			ImageData:    t.ImageData,
			ResponseChan: responseChan,
		}
	}

	// AWAIT RESPONSEEs
	response = transaction.Response{}
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

func (pn *ProcessingReadOnlyNode) Job(t transaction.Transaction) {

	err := pn.Resource.Acquire(context.TODO(), 1)
	if err != nil {
		t.ResponseChan <- transaction.ResponseError(err)
		pn.Resource.Release(1)
		return
	}

	//create reader mirror
	mrr := mirror.NewReader(t.Payload.Reader)

	mirrorPayload := transaction.Payload{
		Reader:     mrr.NewReader(),
		ImageBytes: t.ImageBytes,
	}
	decoded, response := pn.Decode(mirrorPayload, t.ImageData)

	if !response.Ack {
		t.ResponseChan <- response
		pn.Resource.Release(1)
		return
	}

	response = pn.Process(decoded, t.ImageData)

	if !response.Ack {
		t.ResponseChan <- response
		pn.Resource.Release(1)
		return
	}

	pn.Resource.Release(1)

	// SEND
	responseChan := make(chan transaction.Response)
	for _, next := range pn.Next {

		next.Node.GetReceiverChan() <- transaction.Transaction{
			Payload: transaction.Payload{
				Reader:     mrr.NewReader(),
				ImageBytes: t.ImageBytes,
			},
			ImageData:    t.ImageData,
			ResponseChan: responseChan,
		}
	}

	// AWAIT RESPONSEEs
	response = transaction.Response{}
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

func (on *OutputNode) Job(t transaction.Transaction) {
	// TODO assumes output don't have NEXT.
	_ = on.Resource.Acquire(context.TODO(), 1)
	// TODO check err here

	on.TransactionChan() <- t

	on.Resource.Release(1)
}

//////////////

func (n Node) GetReceiverChan() chan transaction.Transaction {
	return n.RecieverChan
}

func (n Node) Job(t transaction.Transaction) {
	panic("virtual")
}
