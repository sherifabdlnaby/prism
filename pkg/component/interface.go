package component

import (
	"go.uber.org/zap"
	"time"
)

// Component defines the basic prism component.
type Component interface {
	// Init Initializes Component's configuration
	Init(Config, zap.SugaredLogger) error

	// start starts the component
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

// Output Component used for outputting data to external destination
type Output interface {
	// TransactionChan returns a channel used to send transactions for saving.
	TransactionChan() chan<- Transaction

	Component
}

//------------------------------------------------------------------------------

//Decoder A Component that can decodes an Input
type Decoder interface {
	Decode(in InputPayload, data ImageData) (interface{}, Response)
}

//Processor A Component that can process an image
type Processor interface {
	Process(in interface{}, data ImageData) (interface{}, Response)
}

//Encoder A Component that encode an Image
type Encoder interface {
	Encode(in interface{}, data ImageData, out *OutputPayload) Response
}

// ProcessorReadWrite can decode, process, or encode a payload.
type ProcessorReadWrite interface {
	Component
	Encoder
	Processor
	Decoder
}

// ProcessorReadOnly can decode, process, and image but doesn't output any data.
type ProcessorReadOnly interface {
	Component
	Processor
	Decoder
}

//------------------------------------------------------------------------------
