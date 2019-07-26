package registry

import (
	"fmt"

	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/component/input"
	"github.com/sherifabdlnaby/prism/pkg/component/output"
	"github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

func (m *Registry) GetComponent(name string) component.Base {
	var component component.Base = nil

	component, ok := m.Inputs[name]
	if !ok {
		component, ok = m.ProcessorReadWrite[name]
		if !ok {
			component, ok = m.ProcessorReadWriteStream[name]
			if !ok {
				component, ok = m.ProcessorReadOnly[name]
				if !ok {
					component, ok = m.Outputs[name]
					if !ok {
						return nil
					}
				}

			}
		}
	}

	return component
}

func (m *Registry) GetProcessorsList() []Processor {
	processorsList := make([]Processor, 0)

	for _, value := range m.ProcessorReadWrite {
		processorsList = append(processorsList, value)
	}

	for _, value := range m.ProcessorReadOnly {
		processorsList = append(processorsList, value)
	}

	for _, value := range m.ProcessorReadWriteStream {
		processorsList = append(processorsList, value)
	}

	return processorsList

}

// --------------------------------------------------------

// LoadInput Load wrapper.Input Plugin in the loaded registry, according to the parsed config.
func (m *Registry) LoadInput(name string, config config.Input, Logger zap.SugaredLogger) error {
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

	m.Inputs[name] = &Input{
		Input:    pluginInstance,
		Resource: *resource.NewResource(config.Concurrency),
	}
	return nil
}

// GetInput Get wrapper.Input Plugin from the loaded plugins.
func (m *Registry) GetInput(name string) *Input {
	return m.Inputs[name]
}

// --------------------------------------------------------

// LoadProcessor Load wrapper.ProcessReadWrite Plugin in the loaded registry, according to the parsed config.
func (m *Registry) LoadProcessor(name string, config config.Processor, Logger zap.SugaredLogger) error {
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
		m.ProcessorReadWrite[name] = &ProcessorReadWrite{
			ReadWrite: plugin,
			Resource:  *resource.NewResource(config.Concurrency),
		}
	case processor.ReadWriteStream:
		m.ProcessorReadWriteStream[name] = &ProcessorReadWriteStream{
			ReadWriteStream: plugin,
			Resource:        *resource.NewResource(config.Concurrency),
		}
	case processor.ReadOnly:
		m.ProcessorReadOnly[name] = &ProcessorReadOnly{
			ReadOnly: plugin,
			Resource: *resource.NewResource(config.Concurrency),
		}
	default:
		return fmt.Errorf("plugin type [%s] is not a processor plugin", config.Plugin)
	}

	return nil
}

// GetProcessor Get wrapper.ProcessReadWrite Plugin from the loaded plugins.
func (m *Registry) GetProcessorReadWrite(name string) (a *ProcessorReadWrite, b bool) {
	a, b = m.ProcessorReadWrite[name]
	return
}

// GetProcessor Get wrapper.ProcessReadWrite Plugin from the loaded plugins.
func (m *Registry) GetProcessorReadWriteStream(name string) (a *ProcessorReadWriteStream, b bool) {
	a, b = m.ProcessorReadWriteStream[name]
	return
}

// GetProcessor Get wrapper.ProcessReadWrite Plugin from the loaded plugins.
func (m *Registry) GetProcessorReadOnly(name string) (a *ProcessorReadOnly, b bool) {
	a, b = m.ProcessorReadOnly[name]
	return
}

// --------------------------------------------------------

// LoadOutput Load wrapper.Output Plugin in the loaded registry, according to the parsed config.
func (m *Registry) LoadOutput(name string, config config.Output, Logger zap.SugaredLogger) error {
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

	txnChan := make(chan transaction.Transaction)
	pluginInstance.SetTransactionChan(txnChan)

	m.Outputs[name] = &Output{
		Output:          pluginInstance,
		Resource:        *resource.NewResource(config.Concurrency),
		TransactionChan: txnChan,
	}

	return nil
}

// GetOutput Get wrapper.Output Plugin from the loaded plugins.
func (m *Registry) GetOutput(name string) *Output {
	return m.Outputs[name]
}

// --------------------------------------------------------

func (m *Registry) exists(name string) bool {
	_, ok := m.Inputs[name]
	if !ok {
		_, ok = m.ProcessorReadWrite[name]
		if !ok {
			_, ok = m.ProcessorReadWriteStream[name]
			if !ok {
				_, ok = m.ProcessorReadOnly[name]
				if !ok {
					_, ok = m.Outputs[name]
				}

			}
		}
	}
	return ok
}
