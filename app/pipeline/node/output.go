package node

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

type Output struct {
	component.Output
}

func (on *Output) Job(t transaction.Transaction) (mirror.ReaderCloner, transaction.ImageBytes, transaction.Response) {
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

	return readerCloner, t.ImageBytes, response
}
