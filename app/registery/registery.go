package registery

import (
	"github.com/sherifabdlnaby/prism/app/registery/wrapper"
	dummyinput "github.com/sherifabdlnaby/prism/internal/input/dummy"
	"github.com/sherifabdlnaby/prism/internal/output/disk"
	dummyprocessor "github.com/sherifabdlnaby/prism/internal/processor/dummy"
	"github.com/sherifabdlnaby/prism/pkg/component"
)

// registered used to key map components names -> constructors.
var registered = map[string]func() component.Component{
	"dummy_processor": dummyprocessor.NewComponent,
	"dummy_input":     dummyinput.NewComponent,
	"disk":            disk.NewComponent,
}

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
