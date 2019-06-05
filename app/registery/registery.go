package registery

import (
	"github.com/sherifabdlnaby/prism/app/registery/wrapper"
)

//Registry Contains Plugin Instances
type Registry struct {
	Inputs     map[string]*wrapper.Input
	Processors map[string]*wrapper.Processor
	Outputs    map[string]*wrapper.Output
}

//NewRegistry Register constructor.
func NewRegistry() *Registry {
	return &Registry{
		Inputs:     make(map[string]*wrapper.Input),
		Processors: make(map[string]*wrapper.Processor),
		Outputs:    make(map[string]*wrapper.Output),
	}
}
