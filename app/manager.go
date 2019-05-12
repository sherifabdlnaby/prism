package app

import (
	"fmt"

	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline"
	componentConfig "github.com/sherifabdlnaby/prism/pkg/config"
)

// loadPlugins Load all plugins in Config
func (a *App) loadPlugins(c config.Config) error {
	a.logger.Info("loading plugins configuration...")

	// Load Input Plugins
	for name, plugin := range c.Inputs.Inputs {
		err := a.registry.LoadInput(name, plugin, a.logger.inputLogger)
		if err != nil {
			return err
		}
	}

	// Load Processor Plugins
	for name, plugin := range c.Processors.Processors {
		err := a.registry.LoadProcessor(name, plugin, a.logger.processingLogger)
		if err != nil {
			return err
		}
	}

	// Load Output Plugins
	for name, plugin := range c.Outputs.Outputs {
		err := a.registry.LoadOutput(name, plugin, a.logger.outputLogger)
		if err != nil {
			return err
		}
	}

	return nil
}

// initPlugins Init all plugins in Config by calling their Init() function
func (a *App) initPlugins(c config.Config) error {
	a.logger.Info("initializing plugins...")

	// Init Input Plugins
	for name, input := range c.Inputs.Inputs {
		plugin, _ := a.registry.GetInput(name)
		pluginConfig := *componentConfig.NewConfig(input.Config)
		err := plugin.Init(pluginConfig, plugin.Logger)
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Processor Plugins
	for name, processor := range c.Processors.Processors {
		plugin, _ := a.registry.GetProcessor(name)
		pluginConfig := *componentConfig.NewConfig(processor.Config)
		err := plugin.Init(pluginConfig, plugin.Logger)
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Output Plugins
	for name, output := range c.Outputs.Outputs {
		plugin, _ := a.registry.GetOutput(name)
		pluginConfig := *componentConfig.NewConfig(output.Config)
		err := plugin.Init(pluginConfig, plugin.Logger)
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	return nil
}

// startInputPlugins startMux all input plugins in Config by calling their startMux() function
func (a *App) startInputPlugins(c config.Config) error {
	a.logger.inputLogger.Info("starting input plugins...")

	for _, value := range a.registry.InputPlugins {
		err := value.Start()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

// startProcessorPlugins startMux all processor plugins in Config by calling their startMux() function
func (a *App) startProcessorPlugins(c config.Config) error {
	a.logger.processingLogger.Info("starting processor plugins...")

	for _, value := range a.registry.ProcessorPlugins {
		err := value.Start()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

// startOutputPlugins startMux all output plugins in Config by calling their startMux() function
func (a *App) startOutputPlugins(c config.Config) error {
	a.logger.Info("starting output plugins...")

	for _, value := range a.registry.OutputPlugins {
		err := value.Start()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

// initPipelines Initialize and build all configured pipelines
func (a *App) initPipelines(c config.Config) error {
	a.logger.Info("initializing pipelines...")

	for key, value := range c.Pipeline.Pipelines {

		// check if pipeline already exists
		_, ok := a.pipelines[key]
		if ok {
			return fmt.Errorf("pipeline with name [%s] already declared", key)
		}

		pip, err := pipeline.NewPipeline(value, a.registry, *a.logger.processingLogger.Named(key))

		if err != nil {
			return fmt.Errorf("error occured when constructing pipeline [%s]: %s", key, err.Error())
		}

		a.pipelines[key] = pip
	}

	return nil
}

// startPipelines startMux all pipelines and start accepting input
func (a *App) startPipelines(c config.Config) error {

	a.logger.Info("starting pipelines...")

	for _, value := range a.pipelines {
		err := value.Start()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

// stopPipelines Stop pipelines by calling their Stop() function, any request to these pipelines will return error.
func (a *App) stopPipelines(c config.Config) error {

	for _, value := range a.pipelines {
		err := value.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

// stopInputPlugins Stop all input plugins in Config by calling their Stop() function
func (a *App) stopInputPlugins(c config.Config) error {

	for _, value := range a.registry.InputPlugins {
		err := value.Close()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

// stopProcessorPlugins Stop all processor plugins in Config by calling their Stop() function
func (a *App) stopProcessorPlugins(c config.Config) error {

	for _, value := range a.registry.ProcessorPlugins {
		err := value.Close()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

// stopOutputPlugins Stop all output plugins in Config by calling their Stop() function
func (a *App) stopOutputPlugins(c config.Config) error {

	for _, value := range a.registry.OutputPlugins {
		err := value.Close()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

//startComponents startMux components
func (a *App) startComponents(c config.Config) error {
	err := a.loadPlugins(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.initPlugins(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.initPipelines(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.startPipelines(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	a.startMux()

	err = a.startOutputPlugins(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.startProcessorPlugins(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.startInputPlugins(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	return nil
}

//stopComponentsGracefully Stop components in graceful strategy
// 		1- Stop Input Components.
// 		2- Stop Pipelines.
// 		3- Stop Processor Components.
// 		4- Stop Output Components.
// As by definition each stop functionality in these components is graceful, this should guarantee graceful shutdown.
func (a *App) stopComponentsGracefully(c config.Config) error {
	a.logger.Info("stopping all components gracefully...")

	///////////////////////////////////////

	a.logger.inputLogger.Info("stopping input plugins...")
	err := a.stopInputPlugins(c)
	if err != nil {
		a.logger.inputLogger.Errorw("failed to stop input plugins", "error", err.Error())
		return err
	}
	a.logger.inputLogger.Info("stopped input plugins successfully.")

	///////////////////////////////////////

	a.logger.pipelineLogger.Info("stopping pipelines...")
	err = a.stopPipelines(c)
	if err != nil {
		a.logger.pipelineLogger.Errorw("failed to stop pipelines", "error", err.Error())
		return err
	}
	a.logger.pipelineLogger.Info("stopping pipelines successfully.")

	///////////////////////////////////////

	err = a.stopProcessorPlugins(c)
	a.logger.processingLogger.Info("stopping processor plugins...")
	if err != nil {
		a.logger.processingLogger.Errorw("failed to stop input plugins", "error", err.Error())
		return err
	}
	a.logger.processingLogger.Info("stopping processor successfully.")

	///////////////////////////////////////

	err = a.stopOutputPlugins(c)
	a.logger.outputLogger.Info("stopping output plugins...")
	if err != nil {
		a.logger.outputLogger.Errorw("failed to stop output plugins", "error", err.Error())
		return err
	}
	a.logger.outputLogger.Info("stopped output successfully.")

	///////////////////////////////////////

	a.logger.Info("stopped all components successfully")

	return nil
}
