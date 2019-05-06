package manager

import (
	"fmt"

	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/mux"
	"github.com/sherifabdlnaby/prism/app/pipeline"
	"github.com/sherifabdlnaby/prism/app/registery"
	componentConfig "github.com/sherifabdlnaby/prism/pkg/config"
)

//Manager Contains all component instances and pipelines, and is responsible for managing them.
type Manager struct {
	logger
	registery.Registry
	Pipelines map[string]*pipeline.Pipeline
	Mux       mux.Mux
}

//NewManager Create a new manager based on the already parsed configs.
func NewManager(c config.Config) *Manager {
	m := Manager{}
	m.logger = *newLoggers(c)
	m.Registry = *registery.NewRegistry()
	m.Pipelines = make(map[string]*pipeline.Pipeline)
	m.Mux = mux.Mux{
		Pipelines: m.Pipelines,
		Inputs:    m.Registry.InputPlugins,
	}
	return &m
}

// loadPlugins Load all plugins in Config
func (m *Manager) loadPlugins(c config.Config) error {
	m.baseLogger.Info("loading plugins configuration...")

	// Load Input Plugins
	for name, plugin := range c.Inputs.Inputs {
		err := m.LoadInput(name, plugin, m.logger.inputLogger)
		if err != nil {
			return err
		}
	}

	// Load Processor Plugins
	for name, plugin := range c.Processors.Processors {
		err := m.LoadProcessor(name, plugin, m.logger.processingLogger)
		if err != nil {
			return err
		}
	}

	// Load Output Plugins
	for name, plugin := range c.Outputs.Outputs {
		err := m.LoadOutput(name, plugin, m.logger.outputLogger)
		if err != nil {
			return err
		}
	}

	return nil
}

// initPlugins Init all plugins in Config by calling their Init() function
func (m *Manager) initPlugins(c config.Config) error {
	m.baseLogger.Info("initializing plugins...")

	// Init Input Plugins
	for name, input := range c.Inputs.Inputs {
		plugin, _ := m.GetInput(name)
		pluginConfig := *componentConfig.NewConfig(input.Config)
		err := plugin.Init(pluginConfig, plugin.Logger)
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Processor Plugins
	for name, processor := range c.Processors.Processors {
		plugin, _ := m.GetProcessor(name)
		pluginConfig := *componentConfig.NewConfig(processor.Config)
		err := plugin.Init(pluginConfig, plugin.Logger)
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Output Plugins
	for name, output := range c.Outputs.Outputs {
		plugin, _ := m.GetOutput(name)
		pluginConfig := *componentConfig.NewConfig(output.Config)
		err := plugin.Init(pluginConfig, plugin.Logger)
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	return nil
}

// startInputPlugins Start all input plugins in Config by calling their Start() function
func (m *Manager) startInputPlugins(c config.Config) error {
	m.inputLogger.Info("starting input plugins...")

	for _, value := range m.InputPlugins {
		err := value.Start()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

// startProcessorPlugins Start all processor plugins in Config by calling their Start() function
func (m *Manager) startProcessorPlugins(c config.Config) error {
	m.processingLogger.Info("starting processor plugins...")

	for _, value := range m.ProcessorPlugins {
		err := value.Start()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

// startOutputPlugins Start all output plugins in Config by calling their Start() function
func (m *Manager) startOutputPlugins(c config.Config) error {
	m.baseLogger.Info("starting output plugins...")

	for _, value := range m.OutputPlugins {
		err := value.Start()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

// initPipelines Initialize and build all configured pipelines
func (m *Manager) initPipelines(c config.Config) error {
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

		m.Pipelines[key] = pip
	}

	return nil
}

// startPipelines Start all pipelines and start accepting input
func (m *Manager) startPipelines(c config.Config) error {

	m.baseLogger.Info("starting pipelines...")

	for _, value := range m.Pipelines {
		err := value.Start()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

//startMux Start the mux that forwards the transactions from input to pipelines based on pipelineTag in transaction.
func (m *Manager) startMux() error {
	m.Mux.Start()
	m.baseLogger.Info("starting forwarding input to pipelines...")
	return nil
}

// stopPipelines Stop pipelines by calling their Stop() function, any request to these pipelines will return error.
func (m *Manager) stopPipelines(c config.Config) error {

	for _, value := range m.Pipelines {
		err := value.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

// stopInputPlugins Stop all input plugins in Config by calling their Stop() function
func (m *Manager) stopInputPlugins(c config.Config) error {

	for _, value := range m.InputPlugins {
		err := value.Close()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

// stopProcessorPlugins Stop all processor plugins in Config by calling their Stop() function
func (m *Manager) stopProcessorPlugins(c config.Config) error {

	for _, value := range m.ProcessorPlugins {
		err := value.Close()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

// stopOutputPlugins Stop all output plugins in Config by calling their Stop() function
func (m *Manager) stopOutputPlugins(c config.Config) error {

	for _, value := range m.OutputPlugins {
		err := value.Close()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

//StartComponents Start components
func (m *Manager) StartComponents(c config.Config) error {
	err := m.loadPlugins(c)
	if err != nil {
		m.baseLogger.Error(err)
		return err
	}

	err = m.initPlugins(c)
	if err != nil {
		m.baseLogger.Error(err)
		return err
	}

	err = m.initPipelines(c)
	if err != nil {
		m.baseLogger.Error(err)
		return err
	}

	err = m.startPipelines(c)
	if err != nil {
		m.baseLogger.Error(err)
		return err
	}

	err = m.startMux()
	if err != nil {
		m.baseLogger.Error(err)
		return err
	}

	err = m.startOutputPlugins(c)
	if err != nil {
		m.baseLogger.Error(err)
		return err
	}

	err = m.startProcessorPlugins(c)
	if err != nil {
		m.baseLogger.Error(err)
		return err
	}

	err = m.startInputPlugins(c)
	if err != nil {
		m.baseLogger.Error(err)
		return err
	}

	return nil
}

//StopComponentsGracefully Stop components in graceful strategy
// 		1- Stop Input Components.
// 		2- Stop Pipelines.
// 		3- Stop Processor Components.
// 		4- Stop Output Components.
// As by definition each stop functionality in these components is graceful, this should guarantee graceful shutdown.
func (m *Manager) StopComponentsGracefully(c config.Config) error {
	m.baseLogger.Info("stopping all components gracefully...")

	///////////////////////////////////////

	m.inputLogger.Info("stopping input plugins...")
	err := m.stopInputPlugins(c)
	if err != nil {
		m.inputLogger.Errorw("failed to stop input plugins", "error", err.Error())
		return err
	}
	m.inputLogger.Info("stopped input plugins successfully.")

	///////////////////////////////////////

	m.pipelineLogger.Info("stopping pipelines...")
	err = m.stopPipelines(c)
	if err != nil {
		m.pipelineLogger.Errorw("failed to stop pipelines", "error", err.Error())
		return err
	}
	m.pipelineLogger.Info("stopping pipelines successfully.")

	///////////////////////////////////////

	err = m.stopProcessorPlugins(c)
	m.processingLogger.Info("stopping processor plugins...")
	if err != nil {
		m.processingLogger.Errorw("failed to stop input plugins", "error", err.Error())
		return err
	}
	m.processingLogger.Info("stopping processor successfully.")

	///////////////////////////////////////

	err = m.stopOutputPlugins(c)
	m.outputLogger.Info("stopping output plugins...")
	if err != nil {
		m.outputLogger.Errorw("failed to stop output plugins", "error", err.Error())
		return err
	}
	m.outputLogger.Info("stopped output successfully.")

	///////////////////////////////////////

	m.baseLogger.Info("stopped all components successfully")

	return nil
}
