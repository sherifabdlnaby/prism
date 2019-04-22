package manager

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/mux"
	"github.com/sherifabdlnaby/prism/app/pipeline"
	"github.com/sherifabdlnaby/prism/app/registery"
	componentConfig "github.com/sherifabdlnaby/prism/pkg/config"
)

type Manager struct {
	logger
	registery.Registry
	Pipelines map[string]pipeline.Pipeline
	Mux       mux.Mux
}

func NewManager(c config.Config) *Manager {
	m := Manager{}
	m.logger = *newLoggers(c)
	m.Registry = *registery.NewRegistry()
	m.Pipelines = make(map[string]pipeline.Pipeline)
	m.Mux = mux.Mux{
		Pipelines: m.Pipelines,
		Inputs:    m.Registry.InputPlugins,
	}
	return &m
}

//LoadPlugins Load all plugins in Config
func (m *Manager) LoadPlugins(c config.Config) error {
	m.baseLogger.Info("loading plugins configuration...")

	// Load Input Plugins
	for name, plugin := range c.Inputs.Inputs {
		err := m.LoadInput(name, plugin)
		if err != nil {
			return err
		}
	}

	// Load Processor Plugins
	for name, plugin := range c.Processors.Processors {
		err := m.LoadProcessor(name, plugin)
		if err != nil {
			return err
		}
	}

	// Load Output Plugins
	for name, plugin := range c.Outputs.Outputs {
		err := m.LoadOutput(name, plugin)
		if err != nil {
			return err
		}
	}

	return nil
}

//InitPlugins Init all plugins in Config by calling their Init() function
func (m *Manager) InitPlugins(c config.Config) error {
	m.baseLogger.Info("initializing plugins...")

	// Init Input Plugins
	for name, input := range c.Inputs.Inputs {
		plugin, _ := m.GetInput(name)
		pluginConfig := *componentConfig.NewConfig(input.Config)
		err := plugin.Init(pluginConfig, *m.inputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Processor Plugins
	for name, processor := range c.Processors.Processors {
		plugin, _ := m.GetProcessor(name)
		pluginConfig := *componentConfig.NewConfig(processor.Config)
		err := plugin.Init(pluginConfig, *m.processingLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Output Plugins
	for name, output := range c.Outputs.Outputs {
		plugin, _ := m.GetOutput(name)
		pluginConfig := *componentConfig.NewConfig(output.Config)
		err := plugin.Init(pluginConfig, *m.outputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	return nil
}

//StartPlugins Start all plugins in Config by calling their Start() function
func (m *Manager) StartPlugins(c config.Config) error {
	m.baseLogger.Info("starting plugins...")

	for _, value := range m.InputPlugins {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	for _, value := range m.ProcessorPlugins {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	for _, value := range m.OutputPlugins {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

//InitPipelines Start all plugins in Config by calling their Start() function
func (m *Manager) InitPipelines(c config.Config) error {
	m.baseLogger.Info("initializing pipelines...")

	for key, value := range c.Pipeline.Pipelines {

		// check if pipeline already exists
		_, ok := m.Pipelines[key]
		if ok {
			return fmt.Errorf("pipeline with name [%s] already declared", key)
		}

		pip, err := pipeline.NewPipeline(value, m.Registry, *m.processingLogger.Named(key))

		if err != nil {
			return fmt.Errorf("error occured when constructing pipeline [%s]: %s", key, err.Error())
		}

		m.Pipelines[key] = *pip
	}

	return nil
}

func (m *Manager) StartPipelines(c config.Config) error {

	m.baseLogger.Info("starting pipelines...")

	for _, value := range m.Pipelines {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) StartMux() error {
	m.baseLogger.Info("starting forwarding input to pipelines...")
	m.Mux.Start()
	return nil
}
