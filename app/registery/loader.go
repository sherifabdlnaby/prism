package registery

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/pkg/component"
)

// LoadInput Load Input Plugin in the loaded registry, according to the parsed config.
func (m *Local) LoadInput(name string, input config.Input) error {
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

	m.InputPlugins[name] = InputWrapper{
		Input:           pluginInstance,
		ResourceManager: *NewResourceManager(input.Concurrency),
	}
	return nil
}

// GetInput Get Input Plugin from the loaded plugins.
func (m *Local) GetInput(name string) (a InputWrapper, b bool) {
	a, b = m.InputPlugins[name]
	return
}

/////////////

// LoadProcessor Load Processor Plugin in the loaded registry, according to the parsed config.
func (m *Local) LoadProcessor(name string, processor config.Processor) error {
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

	m.ProcessorPlugins[name] = ProcessorWrapper{
		ProcessorBase:   pluginInstance,
		ResourceManager: *NewResourceManager(processor.Concurrency),
	}

	return nil
}

// GetProcessor Get Processor Plugin from the loaded plugins.
func (m *Local) GetProcessor(name string) (a ProcessorWrapper, b bool) {
	a, b = m.ProcessorPlugins[name]
	return
}

/////////////

// LoadOutput Load Output Plugin in the loaded registry, according to the parsed config.
func (m *Local) LoadOutput(name string, output config.Output) error {
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

	m.OutputPlugins[name] = OutputWrapper{
		Output:          pluginInstance,
		ResourceManager: *NewResourceManager(output.Concurrency),
	}

	return nil
}

// GetOutput Get Output Plugin from the loaded plugins.
func (m *Local) GetOutput(name string) (a OutputWrapper, b bool) {
	a, b = m.OutputPlugins[name]
	return
}

func (m *Local) exists(name string) bool {
	_, ok := m.InputPlugins[name]
	if !ok {
		_, ok = m.ProcessorPlugins[name]
		if !ok {
			_, ok = m.OutputPlugins[name]
		}
	}
	return ok
}
