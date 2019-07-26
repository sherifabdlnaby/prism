package app

import (
	"fmt"

	"github.com/sherifabdlnaby/prism/app/config"
	componentConfig "github.com/sherifabdlnaby/prism/pkg/config"
)

// loadPlugins Load all plugins in Config
func (a *App) loadPlugins(c config.Config) error {

	// Load Input Plugins
	for name, plugin := range c.Inputs.Inputs {
		err := a.registry.LoadInput(name, *plugin, a.logger.inputLogger)
		if err != nil {
			return err
		}
	}

	// Load Processor Plugins
	for name, plugin := range c.Processors.Processors {
		err := a.registry.LoadProcessor(name, *plugin, a.logger.processingLogger)
		if err != nil {
			return err
		}
	}

	// Load Output Plugins
	for name, plugin := range c.Outputs.Outputs {
		err := a.registry.LoadOutput(name, *plugin, a.logger.outputLogger)
		if err != nil {
			return err
		}
	}

	return nil
}

// initPlugins Init all plugins in Config by calling their Init() function
func (a *App) initPlugins(c config.Config) error {

	// Init Input Plugins
	for name, input := range c.Inputs.Inputs {
		pluginConfig := *componentConfig.NewConfig(input.Config)
		plugin := a.registry.GetInput(name)
		err := plugin.Init(pluginConfig, *a.logger.inputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initialize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Processor Plugins
	for name, processor := range c.Processors.Processors {
		pluginConfig := *componentConfig.NewConfig(processor.Config)
		plugin := a.registry.GetComponent(name)
		err := plugin.Init(pluginConfig, *a.logger.processingLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initialize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Output Plugins
	for name, output := range c.Outputs.Outputs {
		pluginConfig := *componentConfig.NewConfig(output.Config)
		plugin := a.registry.GetOutput(name)
		err := plugin.Init(pluginConfig, *a.logger.outputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initialize plugin [%s]: %s", name, err.Error())
		}
	}

	return nil
}

// startInputPlugins start all input plugins in Config by calling their start() function
func (a *App) startInputPlugins() error {

	for name, value := range a.registry.Inputs {
		err := value.Start()
		if err != nil {
			a.logger.inputLogger.Errorf("failed to start input plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// startProcessorPlugins start all processor plugins in Config by calling their start() function
func (a *App) startProcessorPlugins() error {

	for name, value := range a.registry.GetProcessorsList() {
		err := value.Start()
		if err != nil {
			a.logger.processingLogger.Errorf("failed to start processor plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// startOutputPlugins start all output plugins in Config by calling their start() function
func (a *App) startOutputPlugins() error {

	for name, value := range a.registry.Outputs {
		err := value.Start()
		if err != nil {
			a.logger.outputLogger.Errorf("failed to start output plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// stopInputPlugins Stop all input plugins in Config by calling their Stop() function
func (a *App) stopInputPlugins() error {

	for name, value := range a.registry.Inputs {
		err := value.Close()
		if err != nil {
			a.logger.inputLogger.Errorf("failed to stop input plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// stopProcessorPlugins Stop all processor plugins in Config by calling their Stop() function
func (a *App) stopProcessorPlugins() error {

	for name, value := range a.registry.GetProcessorsList() {
		err := value.Close()
		if err != nil {
			a.logger.processingLogger.Errorf("failed to stop processor plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// stopOutputPlugins Stop all output plugins in Config by calling their Stop() function
func (a *App) stopOutputPlugins() error {

	for name, value := range a.registry.Outputs {
		// close its transaction chan (stop sending txns to it)
		close(value.TransactionChan)

		// close plugin
		err := value.Close()

		if err != nil {
			a.logger.outputLogger.Errorf("failed to stop output plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}
