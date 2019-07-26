package component

import (
	"fmt"

	"github.com/sherifabdlnaby/prism/app/config"
	cfg "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

type logger struct {
	zap.SugaredLogger
	input     zap.SugaredLogger
	output    zap.SugaredLogger
	processor zap.SugaredLogger
}

type Manager struct {
	registry Registry
	logger   logger
}

func (m *Manager) Registry() Registry {
	return m.registry
}

func NewManager(c config.Components, Logger zap.SugaredLogger) (*Manager, error) {
	m := &Manager{
		registry: *NewRegistry(),
		logger: logger{
			SugaredLogger: Logger,
			input:         *Logger.Named("input"),
			processor:     *Logger.Named("processor"),
			output:        *Logger.Named("output"),
		},
	}

	err := m.loadInputs(c.Inputs)
	if err != nil {
		err = fmt.Errorf("error in loading input components, error: %s", err.Error())
		return nil, err
	}

	err = m.loadProcessors(c.Processors)
	if err != nil {
		err = fmt.Errorf("error in loading processors components, error: %s", err.Error())
		return nil, err
	}

	err = m.loadOutputs(c.Outputs)
	if err != nil {
		err = fmt.Errorf("error in loading output components, error: %s", err.Error())
		return nil, err
	}

	return m, nil
}

func (m *Manager) loadInputs(c config.Inputs) error {

	// Load Input Plugins
	for name, component := range c.Inputs {
		err := m.registry.LoadInput(name, *component)
		if err != nil {
			return err
		}

		// INIT
		input := m.registry.Input(name)
		err = input.Init(*cfg.NewConfig(component.Config), *m.logger.input.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initialize component [%s]: %s", name, err.Error())
		}
	}

	return nil
}

func (m *Manager) loadProcessors(c config.Processors) error {

	// Load Input Plugins
	for name, component := range c.Processors {
		err := m.registry.LoadProcessor(name, *component)
		if err != nil {
			return err
		}

		// INIT
		base := m.registry.Processor(name)
		err = base.Init(*cfg.NewConfig(component.Config), *m.logger.processor.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initialize component [%s]: %s", name, err.Error())
		}

	}

	return nil
}

func (m *Manager) loadOutputs(c config.Outputs) error {

	// Load Input Plugins
	for name, component := range c.Outputs {
		err := m.registry.LoadOutput(name, *component)
		if err != nil {
			return err
		}

		// INIT
		output := m.registry.Output(name)
		err = output.Init(*cfg.NewConfig(component.Config), *m.logger.output.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initialize component [%s]: %s", name, err.Error())
		}
	}

	return nil
}

func (m *Manager) StartInput(name string) error {
	input, ok := m.registry.inputs[name]
	if !ok {
		return fmt.Errorf("plugin %s doesn't exist", name)
	}

	err := input.Start()
	if err != nil {
		return fmt.Errorf("failed to start plugin [%s], error: %s", name, err.Error())
	}

	return nil
}

func (m *Manager) StartProcessor(name string) error {
	processor := m.registry.Processor(name)
	if processor == nil {
		return fmt.Errorf("plugin %s doesn't exist", name)
	}

	err := processor.Start()
	if err != nil {
		return fmt.Errorf("failed to start plugin [%s], error: %s", name, err.Error())
	}

	return nil
}

func (m *Manager) StartOutput(name string) error {
	output, ok := m.registry.outputs[name]
	if !ok {
		return fmt.Errorf("plugin %s doesn't exist", name)
	}

	err := output.Start()
	if err != nil {
		return fmt.Errorf("failed to start plugin [%s], error: %s", name, err.Error())
	}

	return nil
}

func (m *Manager) StartAllInputs() error {

	for name := range m.registry.inputs {
		err := m.StartInput(name)
		if err != nil {
			m.logger.input.Errorf("failed to start input plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

func (m *Manager) StartAllOutputs() error {

	for name := range m.registry.outputs {
		err := m.StartOutput(name)
		if err != nil {
			m.logger.output.Errorf("failed to start output plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

func (m *Manager) StartAllProcessors() error {

	for name := range m.registry.AllProcessors() {
		err := m.StartProcessor(name)
		if err != nil {
			m.logger.output.Errorf("failed to start output plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

func (m *Manager) StopInput(name string) error {
	input, ok := m.registry.inputs[name]
	if !ok {
		return fmt.Errorf("plugin %s doesn't exist", name)
	}

	err := input.Stop()
	if err != nil {
		return fmt.Errorf("failed to Stop plugin [%s], error: %s", name, err.Error())
	}

	return nil
}

func (m *Manager) StopProcessor(name string) error {
	processor := m.registry.Processor(name)
	if processor == nil {
		return fmt.Errorf("plugin %s doesn't exist", name)
	}

	err := processor.Stop()
	if err != nil {
		return fmt.Errorf("failed to Stop plugin [%s], error: %s", name, err.Error())
	}

	return nil
}

func (m *Manager) StopOutput(name string) error {
	output, ok := m.registry.outputs[name]
	if !ok {
		return fmt.Errorf("plugin %s doesn't exist", name)
	}

	// close its receive channel
	close(output.TransactionChan)

	err := output.Stop()
	if err != nil {
		return fmt.Errorf("failed to Stop plugin [%s], error: %s", name, err.Error())
	}

	return nil
}

func (m *Manager) StopAllInputs() error {

	for name := range m.registry.inputs {
		err := m.StopInput(name)
		if err != nil {
			m.logger.input.Errorf("failed to Stop input plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

func (m *Manager) StopAllOutputs() error {

	for name := range m.registry.outputs {
		err := m.StopOutput(name)
		if err != nil {
			m.logger.output.Errorf("failed to Stop output plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

func (m *Manager) StopAllProcessors() error {

	for name := range m.registry.AllProcessors() {
		err := m.StopProcessor(name)
		if err != nil {
			m.logger.output.Errorf("failed to Stop output plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

func (m *Manager) InputsReceiveChans() []<-chan transaction.InputTransaction {
	chans := make([]<-chan transaction.InputTransaction, 0)
	for _, in := range m.registry.inputs {
		chans = append(chans, in.InputTransactionChan())
	}
	return chans
}
