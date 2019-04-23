package pipeline

import (
	"context"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

type outputNode struct {
	node
	component.Output
}

func (on *outputNode) Start() {
	go func() {
		for value := range on.RecieverChan {
			go on.Job(value)
		}
	}()
}

func (on *outputNode) Job(t transaction.Transaction) {
	_ = on.Resource.Acquire(context.TODO(), 1)
	// TODO check err here

	responseChan := make(chan transaction.Response)
	readerCloner := mirror.NewReader(t.Payload.Reader)
	mirrorPayload := transaction.Payload{
		Reader:     readerCloner.NewReader(),
		ImageBytes: t.ImageBytes,
	}

	on.TransactionChan() <- transaction.Transaction{
		Payload:      mirrorPayload,
		ImageData:    t.ImageData,
		ResponseChan: responseChan,
	}

	response := <-responseChan

	on.Resource.Release(1)

	if !response.Ack {
		t.ResponseChan <- response
		on.Resource.Release(1)
		return
	}

	// SEND
	on.sendToNextNodes(readerCloner, t.ImageBytes, t.ImageData, responseChan)

	// AWAIT RESPONSEEs
	response = on.receiveResponseFromNextNodes(response, responseChan)

	t.ResponseChan <- response

}
