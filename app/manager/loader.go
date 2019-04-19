package manager

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/pkg/component"
)

var inputPlugins = make(map[string]InputWrapper)
var processorPlugins = make(map[string]ProcessorWrapper)
var outputPlugins = make(map[string]OutputWrapper)

//LoadPlugins Load all plugins in Config
func LoadPlugins(c config.Config) error {
	logger := c.Logger
	logger.Info("loading plugins configuration...")

	// Load Input Plugins
	for name, plugin := range c.Inputs.Inputs {
		err := LoadInput(name, plugin)
		if err != nil {
			return err
		}
	}

	// Load Processor Plugins
	for name, plugin := range c.Processors.Processors {
		err := LoadProcessor(name, plugin)
		if err != nil {
			return err
		}
	}

	// Load Output Plugins
	for name, plugin := range c.Outputs.Outputs {
		err := LoadOutput(name, plugin)
		if err != nil {
			return err
		}
	}

	return nil
}

//InitPlugins Init all plugins in Config by calling their Init() function
func InitPlugins(c config.Config) error {

	logger := c.Logger
	logger.Info("initializing plugins...")

	// Init Input Plugins
	inputLogger := logger.Named("input")
	for name, input := range c.Inputs.Inputs {
		plugin, _ := GetInput(name)
		pluginConfig := *component.NewConfig(input.Config)
		err := plugin.Init(pluginConfig, *inputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Processor Plugins
	processorLogger := logger.Named("processor")
	for name, processor := range c.Processors.Processors {
		plugin, _ := GetProcessor(name)
		pluginConfig := *component.NewConfig(processor.Config)
		err := plugin.Init(pluginConfig, *processorLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Output Plugins
	outputLogger := logger.Named("output")
	for name, output := range c.Outputs.Outputs {
		plugin, _ := GetOutput(name)
		pluginConfig := *component.NewConfig(output.Config)
		err := plugin.Init(pluginConfig, *outputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initalize plugin [%s]: %s", name, err.Error())
		}
	}

	return nil
}

//StartPlugins Start all plugins in Config by calling their Start() function
func StartPlugins(c config.Config) error {

	logger := c.Logger
	logger.Info("starting plugins...")

	for _, value := range inputPlugins {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	for _, value := range processorPlugins {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	for _, value := range outputPlugins {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	return nil
}
