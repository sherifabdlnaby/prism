package registry

import (
	"fmt"

	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/registry/wrapper"
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component/input"
	"github.com/sherifabdlnaby/prism/pkg/component/output"
	"github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

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

	m.Inputs[name] = &wrapper.Input{
		Input:    pluginInstance,
		Resource: *resource.NewResource(config.Concurrency),
	}
	return nil
}

// GetInput Get wrapper.Input Plugin from the loaded plugins.
func (m *Registry) GetInput(name string) (a *wrapper.Input, b bool) {
	a, b = m.Inputs[name]
	return
}

/////////////

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

	switch pluginInstance.(type) {
	case processor.ReadWrite:
	case processor.ReadWriteStream:
	case processor.ReadOnly:
	default:
		return fmt.Errorf("plugin type [%s] is not a processor plugin", config.Plugin)
	}

	m.Processors[name] = &wrapper.Processor{
		Base:     pluginInstance.(processor.Base),
		Resource: *resource.NewResource(config.Concurrency),
	}

	return nil
}

// GetProcessor Get wrapper.ProcessReadWrite Plugin from the loaded plugins.
func (m *Registry) GetProcessor(name string) (a *wrapper.Processor, b bool) {
	a, b = m.Processors[name]
	return
}

/////////////

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

	m.Outputs[name] = &wrapper.Output{
		Output:          pluginInstance,
		Resource:        *resource.NewResource(config.Concurrency),
		TransactionChan: txnChan,
	}

	return nil
}

// GetOutput Get wrapper.Output Plugin from the loaded plugins.
func (m *Registry) GetOutput(name string) (a *wrapper.Output, b bool) {
	a, b = m.Outputs[name]
	return
}

func (m *Registry) exists(name string) bool {
	_, ok := m.Inputs[name]
	if !ok {
		_, ok = m.Processors[name]
		if !ok {
			_, ok = m.Outputs[name]
		}
	}
	return ok
}
