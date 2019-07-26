package component

import (
	"fmt"

	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/pkg/component/input"
	"github.com/sherifabdlnaby/prism/pkg/component/output"
	"github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/job"
)

// LoadInput Load wrapper.Input Plugin in the loaded registry, according to the parsed config.
func (m *Registry) LoadInput(name string, config config.Input) error {
	ok := m.exists(name)
	if ok {
		return fmt.Errorf("duplicate plugin instance with name [%s]", name)
	}

	constructor, ok := registered[config.Plugin]
	if !ok {
		return fmt.Errorf("plugin type [%s] doesn't exist", config.Plugin)
	}

	pluginInstance, ok := constructor().(input.Input)
	if !ok {
		return fmt.Errorf("plugin type [%s] is not an input plugin", config.Plugin)
	}

	m.inputs[name] = &Input{
		Input:    pluginInstance,
		Resource: *NewResource(config.Concurrency),
	}
	return nil
}

// --------------------------------------------------------

// LoadProcessor Load wrapper.ProcessReadWrite Plugin in the loaded registry, according to the parsed config.
func (m *Registry) LoadProcessor(name string, config config.Processor) error {
	ok := m.exists(name)
	if ok {
		return fmt.Errorf("config plugin instance with name [%s] is already loaded", name)
	}

	componentConst, ok := registered[config.Plugin]
	if !ok {
		return fmt.Errorf("config plugin type [%s] doesn't exist", config.Plugin)
	}

	pluginInstance := componentConst()

	switch plugin := pluginInstance.(type) {
	case processor.ReadWrite:
		m.processorReadWrite[name] = &ProcessorReadWrite{
			ReadWrite: plugin,
			Resource:  *NewResource(config.Concurrency),
		}
	case processor.ReadWriteStream:
		m.processorReadWriteStream[name] = &ProcessorReadWriteStream{
			ReadWriteStream: plugin,
			Resource:        *NewResource(config.Concurrency),
		}
	case processor.ReadOnly:
		m.processorReadOnly[name] = &ProcessorReadOnly{
			ReadOnly: plugin,
			Resource: *NewResource(config.Concurrency),
		}
	default:
		return fmt.Errorf("plugin type [%s] is not a processor plugin", config.Plugin)
	}

	return nil
}

// --------------------------------------------------------

// LoadOutput Load wrapper.Output Plugin in the loaded registry, according to the parsed config.
func (m *Registry) LoadOutput(name string, config config.Output) error {
	ok := m.exists(name)
	if ok {
		return fmt.Errorf("config plugin instance with name [%s] is already loaded", name)
	}

	componentConst, ok := registered[config.Plugin]
	if !ok {
		return fmt.Errorf("config plugin type [%s] doesn't exist", config.Plugin)
	}

	pluginInstance, ok := componentConst().(output.Output)
	if !ok {
		return fmt.Errorf("plugin type [%s] is not an output plugin", config.Plugin)
	}

	jobChan := make(chan job.Job)
	pluginInstance.SetJobChan(jobChan)

	m.outputs[name] = &Output{
		Output:   pluginInstance,
		Resource: *NewResource(config.Concurrency),
		JobChan:  jobChan,
	}

	return nil
}

// --------------------------------------------------------

func (m *Registry) exists(name string) bool {
	_, ok := m.inputs[name]
	if !ok {
		_, ok = m.processorReadWrite[name]
		if !ok {
			_, ok = m.processorReadWriteStream[name]
			if !ok {
				_, ok = m.processorReadOnly[name]
				if !ok {
					_, ok = m.outputs[name]
				}

			}
		}
	}
	return ok
}
