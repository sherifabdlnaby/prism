package app

import (
	"fmt"

	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline"
	componentConfig "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

// loadPlugins Load all plugins in Config
func (a *App) loadPlugins(c config.Config) error {

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

	// Init Input Plugins
	for name, input := range c.Inputs.Inputs {
		plugin, _ := a.registry.GetInput(name)
		pluginConfig := *componentConfig.NewConfig(input.Config)
		err := plugin.Init(pluginConfig, *a.logger.inputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Processor Plugins
	for name, processor := range c.Processors.Processors {
		plugin, _ := a.registry.GetProcessor(name)
		pluginConfig := *componentConfig.NewConfig(processor.Config)
		err := plugin.Init(pluginConfig, *a.logger.processingLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Output Plugins
	for name, output := range c.Outputs.Outputs {
		plugin, _ := a.registry.GetOutput(name)
		pluginConfig := *componentConfig.NewConfig(output.Config)
		err := plugin.Init(pluginConfig, *a.logger.outputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	return nil
}

// startInputPlugins startMux all input plugins in Config by calling their startMux() function
func (a *App) startInputPlugins(c config.Config) error {

	for name, value := range a.registry.InputPlugins {
		err := value.Start()
		if err != nil {
			a.logger.inputLogger.Errorf("failed to start input plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// startProcessorPlugins startMux all processor plugins in Config by calling their startMux() function
func (a *App) startProcessorPlugins(c config.Config) error {

	for name, value := range a.registry.ProcessorPlugins {
		err := value.Start()
		if err != nil {
			a.logger.processingLogger.Errorf("failed to start processor plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// startOutputPlugins startMux all output plugins in Config by calling their startMux() function
func (a *App) startOutputPlugins(c config.Config) error {

	for name, value := range a.registry.OutputPlugins {
		err := value.Start()
		if err != nil {
			a.logger.outputLogger.Errorf("failed to start output plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// initPipelines Initialize and build all configured pipelines
func (a *App) initPipelines(c config.Config) error {

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

		tc := make(chan transaction.Transaction)
		pip.SetTransactionChan(tc)

		a.pipelines[key] = Pipeline{
			Pipeline:        *pip,
			TransactionChan: tc,
		}
	}

	return nil
}

// startPipelines startMux all pipelines and start accepting input
func (a *App) startPipelines(c config.Config) error {

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
		// close receiving chan
		close(value.TransactionChan)

		// stop pipeline
		err := value.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

// stopInputPlugins Stop all input plugins in Config by calling their Stop() function
func (a *App) stopInputPlugins(c config.Config) error {

	for name, value := range a.registry.InputPlugins {
		err := value.Close()
		if err != nil {
			a.logger.inputLogger.Errorf("failed to stop input plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// stopProcessorPlugins Stop all processor plugins in Config by calling their Stop() function
func (a *App) stopProcessorPlugins(c config.Config) error {

	for name, value := range a.registry.ProcessorPlugins {
		err := value.Close()
		if err != nil {
			a.logger.processingLogger.Errorf("failed to stop processor plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// stopOutputPlugins Stop all output plugins in Config by calling their Stop() function
func (a *App) stopOutputPlugins(c config.Config) error {

	for name, value := range a.registry.OutputPlugins {
		err := value.Close()
		if err != nil {
			a.logger.outputLogger.Errorf("failed to stop output plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}
