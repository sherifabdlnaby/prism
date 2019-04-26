package node

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//ReadOnly Wraps a readonly component
type ReadOnly struct {
	component.ProcessorReadOnly
}

//Job Process transaction by calling Decode-> Process->
func (pn *ReadOnly) Job(t transaction.Transaction) (mirror.ReaderCloner, transaction.ImageBytes, response.Response) {
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
