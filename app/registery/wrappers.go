package registery

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/semaphore"
)

//ResourceManager contains types required to control access to a resource
type ResourceManager struct {
	*semaphore.Weighted
}

func NewResourceManager(concurrency int) *ResourceManager {
	return &ResourceManager{
		Weighted: semaphore.NewWeighted(int64(concurrency)),
	}
}

// InputWrapper Wraps an Input Plugin Instance
type InputWrapper struct {
	component.Input
	ResourceManager
}

// ProcessorWrapper wraps a processor Plugin Instance
type ProcessorWrapper struct {
	component.ProcessorBase
	ResourceManager
}

// OutputWrapper Wraps and Input Plugin Instance
type OutputWrapper struct {
	component.Output
	ResourceManager
}
