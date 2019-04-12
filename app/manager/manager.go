package manager

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/pkg/types"
	"github.com/sherifabdlnaby/semaphore"
)

/////////////

var InputPlugins = make(map[string]InputWrapper)
var ProcessorPlugins = make(map[string]ProcessorWrapper)
var OutputPlugins = make(map[string]OutputWrapper)

///////////////

type ResourceManager struct {
	semaphore.Weighted
}

func NewResourceManager(concurrency int) *ResourceManager {
	return &ResourceManager{
		Weighted: *semaphore.NewWeighted(int64(concurrency)),
	}
}

//////////////

type InputWrapper struct {
	types.Input
	ResourceManager
}
type ProcessorWrapper struct {
	types.Processor
	ResourceManager
}
type OutputWrapper struct {
	types.Output
	ResourceManager
}

/////////////

func LoadInput(name string, input config.Input) error {
	_, ok := InputPlugins[name]
	if !ok {
		_, ok = ProcessorPlugins[name]
		if !ok {
			_, ok = OutputPlugins[name]
		}
	}

	if ok {
		return fmt.Errorf("duplicate plugin instance with name [%s]", name)
	}

	component, ok := registered[input.Plugin]
	if !ok {
		return fmt.Errorf("plugin type [%s] doesn't exist", input.Plugin)
	}

	pluginInstance, ok := component().(types.Input)

	if !ok {
		return fmt.Errorf("plugin type [%s] is not an input plugin", input.Plugin)
	}

	InputPlugins[name] = InputWrapper{
		Input:           pluginInstance,
		ResourceManager: *NewResourceManager(input.Concurrency),
	}

	return nil
}

func GetInput(name string) (a InputWrapper, b bool) {
	a, b = InputPlugins[name]
	return
}

/////////////

func LoadProcessor(name string, processor config.Processor) error {
	_, ok := InputPlugins[name]
	if !ok {
		_, ok = ProcessorPlugins[name]
		if !ok {
			_, ok = OutputPlugins[name]
		}
	}
	if ok {
		return fmt.Errorf("processor plugin instance with name [%s] is already loaded", name)
	}

	component, ok := registered[processor.Plugin]
	if !ok {
		return fmt.Errorf("processor plugin type [%s] doesn't exist", processor.Plugin)
	}

	pluginInstance, ok := component().(types.Processor)

	if !ok {
		return fmt.Errorf("plugin type [%s] is not a processor plugin", processor.Plugin)
	}

	ProcessorPlugins[name] = ProcessorWrapper{
		Processor:       pluginInstance,
		ResourceManager: *NewResourceManager(processor.Concurrency),
	}

	return nil
}

func GetProcessor(name string) (a ProcessorWrapper, b bool) {
	a, b = ProcessorPlugins[name]
	return
}

/////////////

func LoadOutput(name string, output config.Output) error {
	_, ok := InputPlugins[name]
	if !ok {
		_, ok = ProcessorPlugins[name]
		if !ok {
			_, ok = OutputPlugins[name]
		}
	}
	if ok {
		return fmt.Errorf("output plugin instance with name [%s] is already loaded", name)
	}

	component, ok := registered[output.Plugin]
	if !ok {
		return fmt.Errorf("output plugin type [%s] doesn't exist", output.Plugin)
	}

	pluginInstance, ok := component().(types.Output)

	if !ok {
		return fmt.Errorf("plugin type [%s] is not an output plugin", output.Plugin)
	}

	OutputPlugins[name] = OutputWrapper{
		Output:          pluginInstance,
		ResourceManager: *NewResourceManager(output.Concurrency),
	}

	return nil
}

func GetOutput(name string) (a OutputWrapper, b bool) {
	a, b = OutputPlugins[name]
	return
}
