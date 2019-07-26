package pipeline

import (
	"fmt"

	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline/persistence"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

type wrapper struct {
	*Pipeline
	TransactionChan chan transaction.Transaction
}

type Manager struct {
	pipelines map[string]wrapper
	registry  component.Registry
	logger    zap.SugaredLogger
}

func (m *Manager) Pipelines() map[string]wrapper {
	return m.pipelines
}

// initPipelines Initialize and build all configured pipelines
func NewManager(c config.Pipelines, registry component.Registry, logger zap.SugaredLogger) (*Manager, error) {
	m := Manager{
		pipelines: make(map[string]wrapper),
		registry:  registry,
		logger:    zap.SugaredLogger{},
	}

	m.logger = *logger.Named("pipeline")

	for name, pipConfig := range c.Pipelines {

		// check if pipeline already exists
		_, ok := m.pipelines[name]
		if ok {
			return nil, fmt.Errorf("pipeline with name [%s] already declared", name)
		}

		pip, err := NewPipeline(name, *pipConfig, m.registry, m.logger)
		if err != nil {
			return nil, fmt.Errorf("error occurred when constructing pipeline [%s]: %s", name, err.Error())
		}

		m.pipelines[name] = *pip
	}

	return &m, nil
}

func (m *Manager) PipelinesReceiveChan() map[string]chan<- transaction.Transaction {
	chans := make(map[string]chan<- transaction.Transaction)

	for key, pipeline := range m.pipelines {
		chans[key] = pipeline.TransactionChan
	}

	return chans
}

// startPipelines start all pipelines and start accepting input
func (m *Manager) Start(name string) error {
	var err error

	pipeline, ok := m.pipelines[name]
	if !ok {
		err = fmt.Errorf("pipeline %s doesn't exist", name)
		m.logger.Error(err.Error())
		return err
	}

	err = pipeline.Start()
	if err != nil {
		m.logger.Error(err.Error())
		return err
	}

	return nil
}

// stopPipelines Stop pipelines by calling their Stop() function, any request to these pipelines will return error.
func (m *Manager) Stop(name string) error {
	var err error

	pipeline, ok := m.pipelines[name]
	if !ok {
		err = fmt.Errorf("pipeline %s doesn't exist", name)
		m.logger.Error(err.Error())
		return err
	}

	err = pipeline.Stop()
	if err != nil {
		m.logger.Error(err.Error())
		return err
	}

	return nil
}

// stopPipelines Stop pipelines by calling their Stop() function, any request to these pipelines will return error.
func (m *Manager) Recover(name string) error {
	var err error

	pipeline, ok := m.pipelines[name]
	if !ok {
		err = fmt.Errorf("pipeline %s doesn't exist", name)
		m.logger.Error(err.Error())
		return err
	}

	err = pipeline.recoverAsync()
	if err != nil {
		m.logger.Error(err.Error())
		return err
	}

	return nil
}

// startPipelines start all pipelines and start accepting input
func (m *Manager) StartAll() error {

	for name := range m.pipelines {
		err := m.Start(name)
		if err != nil {
			return err
		}
	}

	return nil
}

// stopPipelines Stop pipelines by calling their Stop() function, any request to these pipelines will return error.
func (m *Manager) StopAll() error {

	for name := range m.pipelines {
		err := m.Stop(name)
		if err != nil {
			return err
		}
	}

	return nil
}

// stopPipelines Stop pipelines by calling their Stop() function, any request to these pipelines will return error.
func (m *Manager) RecoverAsyncAll() error {

	for name := range m.pipelines {
		err := m.Recover(name)
		if err != nil {
			return err
		}
	}

	return nil
}

// stopPipelines Stop pipelines by calling their Stop() function, any request to these pipelines will return error.
func (m *Manager) Cleanup() error {
	persistence.DirectoryCleanup()
	return nil
}
