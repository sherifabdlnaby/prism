package types

import (
	"time"
)

// Component defines the basic prism component.
type Component interface {
	// Init Initializes Component's configuration
	Init(Config) error

	// Start starts the component
	Start() error

	// Stop shutdown down and clean up resources gracefully within a timeout.
	Close(time.Duration) error
}

//------------------------------------------------------------------------------

// Input is a type that sends messages as transactions and waits for a
// response back.
type Input interface {
	// TransactionChan returns a channel used for consuming transactions from
	// this type.
	TransactionChan() <-chan StreamableTransaction

	Component
}

// Processor can decode, process, or encode a payload.
// TODO more documentation ofc
type Processor interface {
	Decode(EncodedPayload) (DecodedPayload, error)

	Process(DecodedPayload) (DecodedPayload, error)

	Encode(DecodedPayload) (EncodedPayload, error)

	Component
}

//------------------------------------------------------------------------------

// Consumer is the higher level consumer type.
type Output interface {
	// TransactionChan returns a channel used to send transactions for saving.
	TransactionChan() chan<- Transaction

	Component
}

//------------------------------------------------------------------------------
