package registry

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

// Processor wraps a processor Plugin Instance
type Processor interface {
	processor.Processor
}

// ProcessorReadOnly wraps a readonly processor Plugin Instance and its resource
type ProcessorReadOnly struct {
	processor.ReadOnly
	Resource resource.Resource
}

// ProcessorReadWrite wraps a read-write processor Plugin Instance and its resource
type ProcessorReadWrite struct {
	processor.ReadWrite
	Resource resource.Resource
}

//ProcessorReadWriteStream wraps a read-write-stream processor Plugin Instance and its resource
type ProcessorReadWriteStream struct {
	processor.ReadWriteStream
	Resource resource.Resource
}

// Output Wraps and Input Plugin Instance
type Output struct {
	output.Output
	Resource        resource.Resource
	TransactionChan chan transaction.Transaction
}
