package manager

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/semaphore"
)

/////////////

///////////////

type resourceManager struct {
	semaphore.Weighted
}

func newResourceManager(concurrency int) *resourceManager {
	return &resourceManager{
		Weighted: *semaphore.NewWeighted(int64(concurrency)),
	}
}

//////////////

// InputWrapper Wraps and Input Plugin Instance
type InputWrapper struct {
	component.Input
	resourceManager
}

// ProcessorWrapper Wraps and Input Plugin Instance
type ProcessorWrapper struct {
	component.ProcessorReadWrite
	resourceManager
}

// OutputWrapper Wraps and Input Plugin Instance
type OutputWrapper struct {
	component.Output
	resourceManager
}

/////////////

// LoadInput Load Input Plugin in the loaded registry, according to the parsed config.
func LoadInput(name string, input config.Input) error {
	ok := exists(name)
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

	inputPlugins[name] = InputWrapper{
		Input:           pluginInstance,
		resourceManager: *newResourceManager(input.Concurrency),
	}

	return nil
}

// GetInput Get Input Plugin from the loaded plugins.
func GetInput(name string) (a InputWrapper, b bool) {
	a, b = inputPlugins[name]
	return
}

/////////////

// LoadProcessor Load Processor Plugin in the loaded registry, according to the parsed config.
func LoadProcessor(name string, processor config.Processor) error {
	ok := exists(name)
	if ok {
		return fmt.Errorf("processor plugin instance with name [%s] is already loaded", name)
	}

	componentConst, ok := registered[processor.Plugin]
	if !ok {
		return fmt.Errorf("processor plugin type [%s] doesn't exist", processor.Plugin)
	}

	pluginInstance, ok := componentConst().(component.ProcessorReadWrite)
	if !ok {
		return fmt.Errorf("plugin type [%s] is not a processor plugin", processor.Plugin)
	}

	processorPlugins[name] = ProcessorWrapper{
		ProcessorReadWrite: pluginInstance,
		resourceManager:    *newResourceManager(processor.Concurrency),
	}

	return nil
}

// GetProcessor Get Processor Plugin from the loaded plugins.
func GetProcessor(name string) (a ProcessorWrapper, b bool) {
	a, b = processorPlugins[name]
	return
}

/////////////

// LoadOutput Load Output Plugin in the loaded registry, according to the parsed config.
func LoadOutput(name string, output config.Output) error {
	ok := exists(name)
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

	outputPlugins[name] = OutputWrapper{
		Output:          pluginInstance,
		resourceManager: *newResourceManager(output.Concurrency),
	}

	return nil
}

// GetOutput Get Output Plugin from the loaded plugins.
func GetOutput(name string) (a OutputWrapper, b bool) {
	a, b = outputPlugins[name]
	return
}

func exists(name string) bool {
	_, ok := inputPlugins[name]
	if !ok {
		_, ok = processorPlugins[name]
		if !ok {
			_, ok = outputPlugins[name]
		}
	}
	return ok
}
