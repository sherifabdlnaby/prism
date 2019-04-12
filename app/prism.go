package app

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/manager"
	"github.com/sherifabdlnaby/prism/pkg/types"
)

// TODO move these to manager

func LoadPlugins(c config.Config) error {
	logger := c.Logger
	logger.Info("loading plugins configuration...")

	// Load Input Plugins
	for name, plugin := range c.Inputs.Inputs {
		err := manager.LoadInput(name, plugin)
		if err != nil {
			return err
		}
	}

	// Load Processor Plugins
	for name, plugin := range c.Processors.Processors {
		err := manager.LoadProcessor(name, plugin)
		if err != nil {
			return err
		}
	}

	// Load Output Plugins
	for name, plugin := range c.Outputs.Outputs {
		err := manager.LoadOutput(name, plugin)
		if err != nil {
			return err
		}
	}

	return nil
}

func InitPlugins(c config.Config) error {

	logger := c.Logger
	logger.Info("initializing plugins...")

	// Init Input Plugins
	inputLogger := logger.Named("input")
	for name, input := range c.Inputs.Inputs {
		plugin, _ := manager.GetInput(name)
		pluginConfig := *types.NewConfig(input.Config)
		err := plugin.Init(pluginConfig, *inputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Processor Plugins
	processorLogger := logger.Named("processor")
	for name, processor := range c.Processors.Processors {
		plugin, _ := manager.GetProcessor(name)
		pluginConfig := *types.NewConfig(processor.Config)
		err := plugin.Init(pluginConfig, *processorLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Output Plugins
	outputLogger := logger.Named("output")
	for name, output := range c.Outputs.Outputs {
		plugin, _ := manager.GetOutput(name)
		pluginConfig := *types.NewConfig(output.Config)
		err := plugin.Init(pluginConfig, *outputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	return nil
}

func StartPlugins(c config.Config) error {

	logger := c.Logger
	logger.Info("starting plugins...")

	for _, value := range manager.InputPlugins {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	for _, value := range manager.ProcessorPlugins {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	for _, value := range manager.OutputPlugins {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	return nil
}
