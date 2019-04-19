package types

// Transaction represent a transaction containing a streamable payload (the message) and a response channel,
// which is used to indicate whether the payload was successfully processed and propagated to the next destinations.
type Transaction struct {
	// InputPayload is the message payload of this transaction.
	InputPayload

	// ResponseChan should receive a response at the end of a transaction,
	// The response itself indicates whether the payload was successfully processed and propagated
	// to the next destinations.
	ResponseChan chan<- Response
}

// Response indicate whether the payload was successfully processed and propagated to the next destinations.
type Response struct {
	// Error is a non-nil error if the payload failed to process.
	Error error

	// Ack indicates that - even though there may not have been an error in
	// processing the payload -, it should not be acknowledged.
	Ack bool
}

var ResponseACK = Response{
	Error: nil,
	Ack:   true,
}

var ResponseNoACK = Response{
	Error: nil,
	Ack:   false,
}

func ResponseError(err error) Response {
	return Response{
		Error: err,
		Ack:   false,
	}
}
