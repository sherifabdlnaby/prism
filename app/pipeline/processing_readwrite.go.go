package pipeline

import (
	"context"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

type processingReadWriteNode struct {
	node
	component.ProcessorReadWrite
}

func (pn *processingReadWriteNode) Start() {
	go func() {
		for value := range pn.RecieverChan {
			go pn.Job(value)
		}
	}()
}

func (pn *processingReadWriteNode) Job(t transaction.Transaction) {

	err := pn.Resource.Acquire(context.TODO(), 1)
	if err != nil {
		t.ResponseChan <- transaction.ResponseError(err)
		pn.Resource.Release(1)
		return
	}

	/// DECODE

	decoded, response := pn.Decode(t.Payload, t.ImageData)

	if !response.Ack {
		t.ResponseChan <- response
		pn.Resource.Release(1)
		return
	}

	/// PROCESS

	decodedPayload, response := pn.Process(decoded, t.ImageData)

	if !response.Ack {
		t.ResponseChan <- response
		pn.Resource.Release(1)
		return
	}

	// ENCODE

	// base Output buffer
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
	pn.sendToNextNodes(&buffer, baseOutput.ImageBytes, t.ImageData, responseChan)

	// AWAIT RESPONSEEs
	response = pn.receiveResponseFromNextNodes(response, responseChan)

	// Send Response back.
	t.ResponseChan <- response
}
