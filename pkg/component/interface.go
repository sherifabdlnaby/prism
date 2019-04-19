package component

import (
	"go.uber.org/zap"
	"time"
)

// Component defines the basic prism component.
type Component interface {
	// Init Initializes Component's configuration
	Init(Config, zap.SugaredLogger) error

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
	TransactionChan() <-chan Transaction

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

type Decoder interface {
	Decode(in InputPayload, data ImageData) (interface{}, Response)
}

type Processor interface {
	Process(in interface{}, data ImageData) (interface{}, Response)
}

type Encoder interface {
	Encode(in interface{}, data ImageData, out *OutputPayload) Response
}

// Processor can decode, process, or encode a payload.
type ProcessorReadWrite interface {
	Component
	Encoder
	Processor
	Decoder
}

type ProcessorReadOnly interface {
	Component
	Processor
	Decoder
}

//------------------------------------------------------------------------------
