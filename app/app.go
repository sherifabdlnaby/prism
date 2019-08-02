package app

import (
	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/forwarder"
	"github.com/sherifabdlnaby/prism/app/pipeline"
)

//App is an self contained instance of Prism app.
type App struct {
	config     config.Config
	logger     logger
	components component.Manager
	pipelines  pipeline.Manager
	forwarder  forwarder.Forwarder
}

//NewApp Construct a new instance of Prism App using parsed config, instance still need to be initialized and started.
func NewApp(config config.Config) (*App, error) {

	app := &App{
		config: config,
		logger: *newLoggers(config),
	}

	components, err := component.NewManager(config.Components, app.logger.SugaredLogger)
	if err != nil {
		return nil, err
	}

	pipelines, err := pipeline.NewManager(config.Pipelines, components.Registry(), app.logger.SugaredLogger)
	if err != nil {
		return nil, err
	}

	forwarder := forwarder.NewForwarder(components.InputsReceiveChans(), pipelines.PipelinesReceiveChan())

	app.components = *components
	app.pipelines = *pipelines
	app.forwarder = *forwarder

	return app, nil
}

//Start Starts the app according to starting strategy
// 1-load plugins
// 2-init plugins
// 3-init pipelines
// 4-start pipelines
// 5-start plugins
func (a *App) Start(config config.Config) error {

	err := a.pipelines.StartAll()
	if err != nil {
		a.logger.Errorf("error while starting pipelines: %v", err)
		return err
	}

	// starting Mux
	a.forwarder.Start()

	a.logger.Info("starting output plugins...")
	err = a.components.StartAllOutputs()
	if err != nil {
		a.logger.Errorf("error while starting output plugins: %v", err)
		return err
	}

	a.logger.Info("starting processor plugins...")
	err = a.components.StartAllProcessors()
	if err != nil {
		a.logger.Errorf("error while starting processor plugins: %v", err)
		return err
	}

	a.logger.Info("checking for any persisted async requests that need to be done...")
	err = a.pipelines.RecoverAsyncAll()
	if err != nil {
		a.logger.Errorf("error while applying persisted async requests: %v", err)
		return err
	}

	a.logger.Info("starting input plugins...")
	err = a.components.StartAllInputs()
	if err != nil {
		a.logger.Errorf("error while starting input plugins: %v", err)
		return err
	}

	a.logger.Info("successfully started all components")
	return nil
}

//Stop Stop all components gracefully, will stop in the following sequence
// 		1- Input Components.
// 		2- Pipelines.
// 		3- Processor Components.
// 		4- Output Components.
func (a *App) Stop() error {
	a.logger.Info("stopping all components gracefully...")

	///////////////////////////////////////

	a.logger.Info("stopping input plugins...")
	err := a.components.StopAllInputs()
	if err != nil {
		a.logger.Errorf("failed to stop input plugins: %v", err)
		return err
	}
	a.logger.Info("stopped input plugins successfully.")

	///////////////////////////////////////

	a.logger.Info("stopping pipelines...")
	err = a.pipelines.StopAll()
	if err != nil {
		a.logger.Errorf("failed to stop pipelines: %v", err)
		return err
	}
	a.logger.Info("stopping pipelines successfully.")

	///////////////////////////////////////

	a.logger.Info("stopping processor plugins...")
	err = a.components.StopAllProcessors()
	if err != nil {
		a.logger.Errorf("failed to stop input plugins: %v", err)
		return err
	}
	a.logger.Info("stopping processor successfully.")

	///////////////////////////////////////

	a.logger.Info("stopping output plugins...")
	err = a.components.StopAllOutputs()
	if err != nil {
		a.logger.Errorf("failed to stop output plugins: %v", err)
		return err
	}
	a.logger.Info("stopped output successfully.")

	///////////////////////////////////////

	a.logger.Info("stopped all components successfully")

	return nil
}
