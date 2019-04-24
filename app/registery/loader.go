package registery

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/registery/wrapper"
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"go.uber.org/zap"
)

// LoadInput Load wrapper.Input Plugin in the loaded registry, according to the parsed config.
func (m *Registry) LoadInput(name string, input config.Input, Logger zap.SugaredLogger) error {
	ok := m.exists(name)
	if ok {
		return fmt.Errorf("duplicate plugin instance with name [%s]", name)
	}

	componentConst, ok := registered[input.Plugin]
	if !ok {
		return fmt.Errorf("plugin type [%s] doesn't exist", input.Plugin)
	}

	pluginInstance, ok := componentConst().(component.Input)
	if !ok {
		return fmt.Errorf("plugin type [%s] is not an input plugin", input.Plugin)
	}

	m.InputPlugins[name] = wrapper.Input{
		Input:    pluginInstance,
		Resource: *resource.NewResource(input.Concurrency, *Logger.Named(name)),
	}
	return nil
}

// GetInput Get wrapper.Input Plugin from the loaded plugins.
func (m *Registry) GetInput(name string) (a wrapper.Input, b bool) {
	a, b = m.InputPlugins[name]
	return
}

/////////////

// LoadProcessor Load wrapper.Processor Plugin in the loaded registry, according to the parsed config.
func (m *Registry) LoadProcessor(name string, processor config.Processor, Logger zap.SugaredLogger) error {
	ok := m.exists(name)
	if ok {
		return fmt.Errorf("processor plugin instance with name [%s] is already loaded", name)
	}

	componentConst, ok := registered[processor.Plugin]
	if !ok {
		return fmt.Errorf("processor plugin type [%s] doesn't exist", processor.Plugin)
	}

	pluginInstance, ok := componentConst().(component.ProcessorBase)
	if !ok {
		return fmt.Errorf("plugin type [%s] is not a processor plugin", processor.Plugin)
	}

	m.ProcessorPlugins[name] = wrapper.Processor{
		ProcessorBase: pluginInstance,
		Resource:      *resource.NewResource(processor.Concurrency, *Logger.Named(name)),
	}

	return nil
}

// GetProcessor Get wrapper.Processor Plugin from the loaded plugins.
func (m *Registry) GetProcessor(name string) (a wrapper.Processor, b bool) {
	a, b = m.ProcessorPlugins[name]
	return
}

/////////////

// LoadOutput Load wrapper.Output Plugin in the loaded registry, according to the parsed config.
func (m *Registry) LoadOutput(name string, output config.Output, Logger zap.SugaredLogger) error {
	ok := m.exists(name)
	if ok {
		return fmt.Errorf("output plugin instance with name [%s] is already loaded", name)
	}

	componentConst, ok := registered[output.Plugin]
	if !ok {
		return fmt.Errorf("output plugin type [%s] doesn't exist", output.Plugin)
	}

	pluginInstance, ok := componentConst().(component.Output)
	if !ok {
		return fmt.Errorf("plugin type [%s] is not an output plugin", output.Plugin)
	}

	m.OutputPlugins[name] = wrapper.Output{
		Output:   pluginInstance,
		Resource: *resource.NewResource(output.Concurrency, *Logger.Named(name)),
	}

	return nil
}

// GetOutput Get wrapper.Output Plugin from the loaded plugins.
func (m *Registry) GetOutput(name string) (a wrapper.Output, b bool) {
	a, b = m.OutputPlugins[name]
	return
}

func (m *Registry) exists(name string) bool {
	_, ok := m.InputPlugins[name]
	if !ok {
		_, ok = m.ProcessorPlugins[name]
		if !ok {
			_, ok = m.OutputPlugins[name]
		}
	}
	return ok
}
