package component

import (
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

// Component defines the basic prism component.
type Component interface {
	// Init Initializes Component's configuration
	Init(config.Config, zap.SugaredLogger) error

	// start starts the component
	Start() error

	// Stop shutdown down and clean up resources gracefully within a timeout.
	Close() error
}

//------------------------------------------------------------------------------

// Input is a type that sends messages as transactions and waits for a
// response back.
type Input interface {
	// TransactionChan returns a channel used for consuming transactions from
	// this type.
	TransactionChan() <-chan transaction.InputTransaction

	Component
}

//------------------------------------------------------------------------------

// Output Component used for outputting data to external destination
type Output interface {
	// TransactionChan returns a channel used to send transactions for saving.
	TransactionChan() chan<- transaction.Transaction

	Component
}

//------------------------------------------------------------------------------

//Decoder A Component that can decodes an Input
type Decoder interface {
	Decode(in transaction.Payload, data transaction.ImageData) (interface{}, response.Response)
}

//Processor A Component that can process an image
type Processor interface {
	Process(in interface{}, data transaction.ImageData) (interface{}, response.Response)
}

//ProcessorRead A base component that can process images in read-only mode (no-output)
type ProcessorRead interface {
	Process(in interface{}, data transaction.ImageData) response.Response
}

//Encoder A Component that encode an Image
type Encoder interface {
	Encode(in interface{}, data transaction.ImageData, out *transaction.OutputPayload) response.Response
}

// ProcessorBase can process a payload.
type ProcessorBase interface {
	Component
	Decoder
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
	ProcessorRead
	Decoder
}

//------------------------------------------------------------------------------
