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

//Start Start all components configured in the yaml files.
func (a *App) Start(config config.Config) error {
	return a.manager.StartComponents(config)
}

//Stop Stop all components gracefully, will stop in the following sequence
// 		1- Input Components.
// 		2- Pipelines.
// 		3- Processor Components.
// 		4- Output Components.
func (a *App) Stop(config config.Config) error {
	return a.manager.StopComponentsGracefully(config)
}
