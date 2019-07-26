package component

import (
	"github.com/sherifabdlnaby/prism/pkg/component/input"
	"github.com/sherifabdlnaby/prism/pkg/component/output"
	"github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/job"
)

// Input Wraps an Input Plugin Instance
type Input struct {
	input.Input
	Resource Resource
}

// Processor wraps a processor Plugin Instance
type Processor interface {
	processor.Processor
}

// processorReadOnly wraps a readonly processor Plugin Instance and its resource
type ProcessorReadOnly struct {
	processor.ReadOnly
	Resource Resource
}

// processorReadWrite wraps a read-write processor Plugin Instance and its resource
type ProcessorReadWrite struct {
	processor.ReadWrite
	Resource Resource
}

//processorReadWriteStream wraps a read-write-stream processor Plugin Instance and its resource
type ProcessorReadWriteStream struct {
	processor.ReadWriteStream
	Resource Resource
}

// Output Wraps and Input Plugin Instance
type Output struct {
	output.Output
	Resource Resource
	JobChan  chan job.Job
}
