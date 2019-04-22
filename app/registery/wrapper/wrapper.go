package wrapper

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/semaphore"
)

//Resource contains types required to control access to a resource
type Resource struct {
	*semaphore.Weighted
}

func NewResource(concurrency int) *Resource {
	return &Resource{
		Weighted: semaphore.NewWeighted(int64(concurrency)),
	}
}

// Input Wraps an Input Plugin Instance
type Input struct {
	component.Input
	Resource
}

// Processor wraps a processor Plugin Instance
type Processor struct {
	component.ProcessorBase
	Resource
}

// Output Wraps and Input Plugin Instance
type Output struct {
	component.Output
	Resource
}
