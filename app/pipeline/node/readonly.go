package node

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

type ReadOnly struct {
	component.ProcessorReadOnly
}

func (pn *ReadOnly) Job(t transaction.Transaction) (mirror.ReaderCloner, transaction.ImageBytes, transaction.Response) {
	//create reader mirror
	readerCloner := mirror.NewReader(t.Payload.Reader)
	mirrorPayload := transaction.Payload{
		Reader:     readerCloner.NewReader(),
		ImageBytes: t.ImageBytes,
	}

	/// DECODE
	decoded, response := pn.Decode(mirrorPayload, t.ImageData)
	if !response.Ack {
		return nil, nil, response
	}

	response = pn.Process(decoded, t.ImageData)

	return readerCloner, t.ImageBytes, response
}
