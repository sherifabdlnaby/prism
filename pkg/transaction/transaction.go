package transaction

import (
	"context"

	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
)

// Transaction represent a transaction containing a streamable payload (the message) and a response channel,
// which is used to indicate whether the payload was successfully processed and propagated to the next destinations.
type Transaction struct {
	// Payload is the message payload of this transaction.
	Payload payload.Payload

	// Data is the message data of this transaction.
	Data payload.Data

	// Context of the transaction
	Context context.Context

	// ResponseChan should receive a response at the end of a transaction,
	// The response itself indicates whether the payload was successfully processed and propagated
	// to the next destinations.
	ResponseChan chan<- response.Response
}

// InputTransaction represent a transaction containing a streamable payload, Data, PipelineTag, and a response channel,
// response indicate whether the payload was successfully processed and propagated to the next destinations.
// PipelineTag indicate to which pipeline should this transaction be forwarded to.
type InputTransaction struct {
	Transaction
	PipelineTag string
}
