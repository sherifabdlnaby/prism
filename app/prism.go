package app

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/manager"
	"github.com/sherifabdlnaby/prism/pkg/types"
)

func LoadPlugins(c config.Config) error {

	// Load Input Plugins
	for name, plugin := range c.Inputs.Inputs {
		err := manager.Load(name, plugin.Plugin)
		if err != nil {
			return err
		}

		// check if plugin is of type Input
		_, ok := manager.Get(name).(types.Input)
		if !ok {
			return fmt.Errorf("plugin [%s] of type [%s] is not an input plugin", name, plugin.Plugin)
		}
	}

	// Load Processor Plugins
	for name, plugin := range c.Processors.Processors {
		err := manager.Load(name, plugin.Plugin)
		if err != nil {
			return err
		}

		// check if plugin is of type Input
		_, ok := manager.Get(name).(types.Processor)
		if !ok {
			return fmt.Errorf("plugin [%s] of type [%s] is not a processor plugin", name, plugin.Plugin)
		}
	}

	// Load Output Plugins
	for name, plugin := range c.Outputs.Outputs {
		err := manager.Load(name, plugin.Plugin)
		if err != nil {
			return err
		}

		// check if plugin is of type Input
		_, ok := manager.Get(name).(types.Output)
		if !ok {
			return fmt.Errorf("plugin [%s] of type [%s] is not an output plugin", name, plugin.Plugin)
		}
	}

	return nil
}

func InitPlugins(c config.Config) error {

	logger := c.Logger

	// Init Input Plugins
	inputLogger := logger.Named("input")
	for name, input := range c.Inputs.Inputs {
		plugin := manager.Get(name)
		pluginConfig := *types.NewConfig(input.Config)
		err := plugin.Init(pluginConfig, *inputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Processor Plugins
	processorLogger := logger.Named("processor")
	for name, processor := range c.Processors.Processors {
		plugin := manager.Get(name)
		pluginConfig := *types.NewConfig(processor.Config)
		err := plugin.Init(pluginConfig, *processorLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Output Plugins
	outputLogger := logger.Named("output")
	for name, output := range c.Outputs.Outputs {
		plugin := manager.Get(name)
		pluginConfig := *types.NewConfig(output.Config)
		err := plugin.Init(pluginConfig, *outputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	return nil
}
