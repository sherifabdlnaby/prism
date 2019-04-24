package node

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

type ReadWrite struct {
	component.ProcessorReadWrite
}

func (pn *ReadWrite) Job(t transaction.Transaction) (mirror.ReaderCloner, transaction.ImageBytes, transaction.Response) {

	/// DECODE

	decoded, response := pn.Decode(t.Payload, t.ImageData)

	if !response.Ack {
		return nil, nil, response
	}

	/// PROCESS

	decodedPayload, response := pn.Process(decoded, t.ImageData)

	if !response.Ack {
		return nil, nil, response
	}

	// base Output buffer
	buffer := mirror.Writer{}
	baseOutput := transaction.OutputPayload{
		WriteCloser: &buffer,
		ImageBytes:  nil,
	}

	// ENCODE
	response = pn.Encode(decodedPayload, t.ImageData, &baseOutput)

	return &buffer, baseOutput.ImageBytes, response
}
