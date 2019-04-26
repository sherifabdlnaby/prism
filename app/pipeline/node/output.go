package node

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//Output Wraps an output component
type Output struct {
	component.Output
}

//Job Output Job will send the transaction to output plugin and await its result.
func (on *Output) Job(t transaction.Transaction) (mirror.ReaderCloner, transaction.ImageBytes, response.Response) {
	responseChan := make(chan response.Response)
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
