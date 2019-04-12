package pipeline

import (
	"context"
	"github.com/sherifabdlnaby/prism/app/manager"
	"github.com/sherifabdlnaby/prism/pkg/types"
)

//TODO refactor to make output and process close to each other.

type Node struct {
	RecieverChan chan types.Transaction
	Next         []NextNode
}

type NextNode struct {
	async bool
	Node  NodeInterface
}

func (n Node) GetRecieverChan() chan types.Transaction {
	return n.RecieverChan
}

type NodeInterface interface {
	Start()
	GetRecieverChan() chan types.Transaction
	Job(t types.Transaction)
}

///////////////

type DummyNode struct {
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

func (dn *DummyNode) Start() {
	go func() {
		for value := range dn.RecieverChan {
			go dn.Job(value)
		}
	}()
}

func (dn *DummyNode) Job(t types.Transaction) {

	// SEND
	responseChan := make(chan types.Response)
	for _, next := range dn.Next {

		next.Node.GetRecieverChan() <- types.Transaction{
			Payload:      t.Payload,
			ResponseChan: responseChan,
		}
	}

	count, total := 0, len(dn.Next)

	// AWAIT RESPONSEEs
	response := types.Response{}
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

func (pn *ProcessingNode) Start() {
	go func() {
		for value := range pn.RecieverChan {
			go pn.Job(value)
		}
	}()
}

func (pn *ProcessingNode) Job(t types.Transaction) {

	err := pn.ProcessorWrapper.Acquire(context.TODO(), 1)
	// TODO check err here

	decoded, err := pn.Decode(t.Payload)

	if err != nil {
		t.ResponseChan <- types.ResponseError(err)
		pn.ProcessorWrapper.Release(1)
		return
	}

	decodedPayload, err := pn.Process(decoded)

	if err != nil {
		t.ResponseChan <- types.ResponseError(err)
		pn.ProcessorWrapper.Release(1)
		return
	}

	encoded, err := pn.Encode(decodedPayload)

	if err != nil {
		t.ResponseChan <- types.ResponseError(err)
		pn.ProcessorWrapper.Release(1)
		return
	}

	pn.ProcessorWrapper.Release(1)

	// SEND
	responseChan := make(chan types.Response)
	for _, next := range pn.Next {

		next.Node.GetRecieverChan() <- types.Transaction{
			Payload:      encoded,
			ResponseChan: responseChan,
		}
	}

	count, total := 0, len(pn.Next)

	// AWAIT RESPONSEEs
	response := types.Response{}
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

func (pn *OutputNode) Job(t types.Transaction) {
	// TODO assumes output don't have NEXT.
	_ = pn.OutputWrapper.Acquire(context.TODO(), 1)
	// TODO check err here

	pn.TransactionChan() <- t

	pn.OutputWrapper.Release(1)
}
