package wrapper

import (
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component/input"
	"github.com/sherifabdlnaby/prism/pkg/component/output"
	"github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

// Input Wraps an Input Plugin Instance
type Input struct {
	input.Input
	Resource resource.Resource
}

// ProcessReadWrite wraps a processor Plugin Instance
type Processor struct {
	processor.Base
	Resource resource.Resource
}

// ProcessReadWrite wraps a processor Plugin Instance
type ProcessorReadOnly struct {
	processor.ReadOnly
	Resource resource.Resource
}

// ProcessReadWrite wraps a processor Plugin Instance
type ProcessorReadWrite struct {
	processor.ReadWrite
	Resource resource.Resource
}

type ProcessorReadWriteStream struct {
	processor.ReadWriteStream
	Resource resource.Resource
}

// Output Wraps and Input Plugin Instance
type Output struct {
	output.Output
	Resource              resource.Resource
	TransactionChan       chan transaction.Transaction
	StreamTransactionChan chan transaction.Streamable
}
