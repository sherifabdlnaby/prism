package pipeline

import (
	"fmt"
	"time"

	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline/persistence"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"go.uber.org/zap"
)

type wrapper struct {
	*pipeline
	jobChan chan job.Job
}

type Manager struct {
	pipelines   map[string]wrapper
	persistence persistence.Repository
	registry    component.Registry
	logger      zap.SugaredLogger
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

	repo, err := persistence.NewRepository(config.EnvPrismDataDir.Lookup(), m.logger)
	if err != nil {
		return nil, fmt.Errorf("error occurred when constructing pipeline persistence: %s", err.Error())
	}
	m.persistence = *repo

	for name, pipConfig := range c.Pipelines {

		// check if pipeline already exists
		_, ok := m.pipelines[name]
		if ok {
			return nil, fmt.Errorf("pipeline with name [%s] already declared", name)
		}

		pip, err := m.NewPipeline(name, *pipConfig)
		if err != nil {
			return nil, fmt.Errorf("error occurred when constructing pipeline [%s]: %s", name, err.Error())
		}

		m.pipelines[name] = *pip
	}

	return &m, nil
}

func (m *Manager) PipelinesReceiveChan() map[string]chan<- job.Job {
	chans := make(map[string]chan<- job.Job)

	for key, pipeline := range m.pipelines {
		chans[key] = pipeline.jobChan
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

	errChan := make(chan error)
	go func() {

		close(pipeline.jobChan)

		err = pipeline.Stop()
		if err != nil {
			m.logger.Error(err.Error())
		}

		errChan <- err
	}()

	for {
		select {
		case <-time.Tick(50 * time.Millisecond):
			// Print how many active job for visibility
			m.logger.Infof("stopping pipeline [%s]... (jobs in progress: %d)", pipeline.name, pipeline.ActiveJobs())
		case err := <-errChan:
			return err
		}

	}
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

	err = pipeline.recoverAsyncJobs()
	if err != nil {
		m.logger.Error(err.Error())
		return err
	}

	return nil
}

// startPipelines start all pipelines and start accepting input
func (m *Manager) StartAll() error {
	errChan := make(chan error)

	// stop pipelines concurrently
	for name := range m.pipelines {
		go func(name string) {
			err := m.Start(name)
			errChan <- err
		}(name)
	}

	// wait for errors
	var err error
	for i := 0; i < len(m.pipelines); i++ {
		err1 := <-errChan
		if err1 != nil {
			err = err1
		}
	}

	return err
}

// stopPipelines Stop pipelines by calling their Stop() function, any request to these pipelines will return error.
func (m *Manager) StopAll() error {

	errChan := make(chan error)

	// stop pipelines concurrently
	for name := range m.pipelines {
		go func(name string) {
			err := m.Stop(name)
			errChan <- err
		}(name)
	}

	// wait for errors
	var err error
	for i := 0; i < len(m.pipelines); i++ {
		err1 := <-errChan
		if err1 != nil {
			err = err1
		}
	}

	return err
}

// stopPipelines Stop pipelines by calling their Stop() function, any request to these pipelines will return error.
func (m *Manager) RecoverAsyncAll() error {
	errChan := make(chan error)

	// recover pipelines concurrently
	for name := range m.pipelines {
		go func(name string) {
			err := m.Recover(name)
			errChan <- err
		}(name)
	}

	// wait for errors
	var err error
	for i := 0; i < len(m.pipelines); i++ {
		err1 := <-errChan
		if err1 != nil {
			err = err1
		}
	}

	return err
}
