package pipeline

import (
	"context"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

type processingReadOnlyNode struct {
	node
	component.ProcessorReadOnly
}

func (pn *processingReadOnlyNode) Start() {
	go func() {
		for value := range pn.RecieverChan {
			go pn.Job(value)
		}
	}()
}

func (pn *processingReadOnlyNode) Job(t transaction.Transaction) {

	err := pn.Resource.Acquire(context.TODO(), 1)
	if err != nil {
		t.ResponseChan <- transaction.ResponseError(err)
		pn.Resource.Release(1)
		return
	}

	//create reader mirror
	readerCloner := mirror.NewReader(t.Payload.Reader)
	mirrorPayload := transaction.Payload{
		Reader:     readerCloner.NewReader(),
		ImageBytes: t.ImageBytes,
	}

	/// DECODE
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
	pn.sendToNextNodes(readerCloner, t.ImageBytes, t.ImageData, responseChan)

	// AWAIT RESPONSEEs
	response = pn.receiveResponseFromNextNodes(response, responseChan)

	// Send Response back.
	t.ResponseChan <- response
}
