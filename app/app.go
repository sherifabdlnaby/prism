package app

import (
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/manager"
)

//App is an self contained instance of Prism app.
type App struct {
	config  config.Config
	manager manager.Manager
}

//NewApp Construct a new instance of Prism App using parsed config, instance still need to be initialized and started.
func NewApp(config config.Config) *App {
	return &App{
		config:  config,
		manager: *manager.NewManager(config),
	}
}

//InitializeComponents Initialize components in all yaml files, this is responsible for validating the config syntax.
func (a *App) InitializeComponents(config config.Config) error {
	err := a.manager.LoadPlugins(config)
	if err != nil {
		config.Logger.Panic(err)
	}

	err = a.manager.InitPlugins(config)
	if err != nil {
		config.Logger.Panic(err)
	}

	return err
}

//StartComponents Start all components configured in the yaml files.
func (a *App) StartComponents(config config.Config) error {
	err := a.manager.StartOutputPlugins(config)
	if err != nil {
		config.Logger.Panic(err)
	}

	err = a.manager.StartProcessorPlugins(config)
	if err != nil {
		config.Logger.Panic(err)
	}

	err = a.manager.StartInputPlugins(config)
	if err != nil {
		config.Logger.Panic(err)
	}

	return err
}

//InitializePipelines Initialize Pipelines according to config files.
func (a *App) InitializePipelines(config config.Config) error {
	err := a.manager.InitPipelines(config)

	if err != nil {
		config.Logger.Panic(err)
	}

	return err
}

//StartPipelines Start pipelines ( start accepting transactions )
func (a *App) StartPipelines(config config.Config) error {
	err := a.manager.StartPipelines(config)

	if err != nil {
		config.Logger.Panic(err)
	}

	return err
}

//StartMux Start the mux that forwards the transactions from input to pipelines based on pipelineTag in transaction.
func (a *App) StartMux(config config.Config) error {
	err := a.manager.StartMux()

	if err != nil {
		config.Logger.Panic(err)
	}

	return err
}

////

//StopComponents Stop all components gracefully, will stop in the following sequence
// 		1- Input Components.
// 		2- Pipelines.
// 		3- Processor Components.
// 		4- Output Components.
func (a *App) StopComponents(config config.Config) error {
	return a.manager.StopComponentsGracefully(config)
}
