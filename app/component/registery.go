package component

import "github.com/sherifabdlnaby/prism/pkg/component"

//Registry Contains Plugin Instances
type Registry struct {
	inputs                   map[string]*Input
	processorReadOnly        map[string]*ProcessorReadOnly
	processorReadWrite       map[string]*ProcessorReadWrite
	processorReadWriteStream map[string]*ProcessorReadWriteStream
	outputs                  map[string]*Output
}

//NewRegistry Register constructor.
func NewRegistry() *Registry {
	return &Registry{
		inputs:                   make(map[string]*Input),
		processorReadOnly:        make(map[string]*ProcessorReadOnly),
		processorReadWrite:       make(map[string]*ProcessorReadWrite),
		processorReadWriteStream: make(map[string]*ProcessorReadWriteStream),
		outputs:                  make(map[string]*Output),
	}
}

func (m *Registry) Component(key string) component.Base {
	var component component.Base = nil

	component, ok := m.processorReadWrite[key]
	if !ok {
		component, ok = m.processorReadWriteStream[key]
		if !ok {
			component, ok = m.processorReadOnly[key]
			if !ok {
				component, ok = m.outputs[key]
				if !ok {
					return nil
				}
			}

		}
	}

	return component
}

func (m *Registry) AllProcessors() map[string]Processor {
	processors := make(map[string]Processor)

	// It's safe to not worry about key conflicts as they're checked before being added by exists()

	for k, v := range m.processorReadWrite {
		processors[k] = v
	}

	for k, v := range m.processorReadOnly {
		processors[k] = v
	}

	for k, v := range m.processorReadWriteStream {
		processors[k] = v
	}

	return processors
}

func (m *Registry) Processor(key string) Processor {

	// It's safe to not worry about key conflicts as they're checked before being added by exists()

	for k, v := range m.processorReadWrite {
		if key == k {
			return v
		}
	}

	for k, v := range m.processorReadOnly {
		if key == k {
			return v
		}
	}

	for k, v := range m.processorReadWriteStream {
		if key == k {
			return v
		}
	}

	return nil

}

// Input Get wrapper.Input Plugin from the loaded plugins.
func (m *Registry) Input(name string) *Input {
	return m.inputs[name]
}

// Output Get wrapper.Output Plugin from the loaded plugins.
func (m *Registry) Output(name string) *Output {
	return m.outputs[name]
}
