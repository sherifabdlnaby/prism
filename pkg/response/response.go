package response

// Response indicate whether the payload was successfully processed and propagated to the next destinations.
type Response struct {
	// Error is a non-nil error if the payload failed to process due to an *internal* error.
	Error error

	// Ack indicates that - even though there may not have been an error in
	// processing the payload -, it should not be acknowledged.
	Ack bool

	// AckErr is why the
	AckErr error
}

//ACK A Successful Ack
var ACK = Response{
	Error:  nil,
	Ack:    true,
	AckErr: nil,
}

//Error Return a no-ack response that indicated an *internal* error due to a failure.
func Error(err error) Response {
	return Response{
		Error: err,
		Ack:   false,
	}
}

// NoAck Return a no-ack response that indicates that the message should be not-acknowledged, not due to a failure.
// reason describes *why* the message was not-acknowledged.
func NoAck(reason error) Response {
	return Response{
		Error:  nil,
		Ack:    false,
		AckErr: reason,
	}
}

//Ack returns A Successful Ack response
func Ack() Response {
	return ACK
}
