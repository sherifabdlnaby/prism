package registry

import (
	"github.com/sherifabdlnaby/prism/app/registry/wrapper"
)

//Registry Contains Plugin Instances
type Registry struct {
	Inputs                   map[string]*wrapper.Input
	ProcessorReadOnly        map[string]*wrapper.ProcessorReadOnly
	ProcessorReadWrite       map[string]*wrapper.ProcessorReadWrite
	ProcessorReadWriteStream map[string]*wrapper.ProcessorReadWriteStream
	Outputs                  map[string]*wrapper.Output
}

//NewRegistry Register constructor.
func NewRegistry() *Registry {
	return &Registry{
		Inputs:                   make(map[string]*wrapper.Input),
		ProcessorReadOnly:        make(map[string]*wrapper.ProcessorReadOnly),
		ProcessorReadWrite:       make(map[string]*wrapper.ProcessorReadWrite),
		ProcessorReadWriteStream: make(map[string]*wrapper.ProcessorReadWriteStream),
		Outputs:                  make(map[string]*wrapper.Output),
	}
}
