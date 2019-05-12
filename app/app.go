package app

import (
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline"
	"github.com/sherifabdlnaby/prism/app/registery"
)

//App is an self contained instance of Prism app.
type App struct {
	config    config.Config
	logger    logger
	registry  registery.Registry
	pipelines map[string]*pipeline.Pipeline
}

//NewApp Construct a new instance of Prism App using parsed config, instance still need to be initialized and started.
func NewApp(config config.Config) *App {

	app := &App{
		config:    config,
		logger:    *newLoggers(config),
		registry:  *registery.NewRegistry(),
		pipelines: make(map[string]*pipeline.Pipeline),
	}

	return app
}

//startMux startMux all components configured in the yaml files.
func (a *App) Start(config config.Config) error {
	return a.startComponents(config)
}

//Stop Stop all components gracefully, will stop in the following sequence
// 		1- Input Components.
// 		2- Pipelines.
// 		3- Processor Components.
// 		4- Output Components.
func (a *App) Stop(config config.Config) error {
	return a.stopComponentsGracefully(config)
}
