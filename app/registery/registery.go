package registery

import (
	"github.com/sherifabdlnaby/prism/app/registery/wrapper"
)

//Registry Contains Plugin Instances
type Registry struct {
	InputPlugins     map[string]*wrapper.Input
	ProcessorPlugins map[string]*wrapper.Processor
	OutputPlugins    map[string]*wrapper.Output
}

//NewRegistry Register constructor.
func NewRegistry() *Registry {
	return &Registry{
		InputPlugins:     make(map[string]*wrapper.Input),
		ProcessorPlugins: make(map[string]*wrapper.Processor),
		OutputPlugins:    make(map[string]*wrapper.Output),
	}
}
