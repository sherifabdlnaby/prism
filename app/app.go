package app

import (
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline"
	"github.com/sherifabdlnaby/prism/app/registry"
)

//App is an self contained instance of Prism app.
type App struct {
	config         config.Config
	logger         logger
	registry       registry.Registry
	pipelineManger pipeline.Manager
}

//NewApp Construct a new instance of Prism App using parsed config, instance still need to be initialized and started.
func NewApp(config config.Config) (*App, error) {

	app := &App{
		config:   config,
		logger:   *newLoggers(config),
		registry: *registry.NewRegistry(), //TODO delete when create component manager
	}

	pipelineManager, err := pipeline.NewManager(config.Pipeline, app.logger.SugaredLogger)
	if err != nil {
		return nil, err
	}

	app.pipelineManger = *pipelineManager

	return app, nil
}

//Start Starts the app according to starting strategy
// 1-load plugins
// 2-init plugins
// 3-init pipelines
// 4-start pipelines
// 5-start plugins
func (a *App) Start(config config.Config) error {
	return a.startComponents(config)
}

//Stop Stop all components gracefully, will stop in the following sequence
// 		1- Input Components.
// 		2- Pipelines.
// 		3- Processor Components.
// 		4- Output Components.
func (a *App) Stop() error {
	return a.stopComponents()
}

//startComponents start components
func (a *App) startComponents(c config.Config) error {

	a.logger.Info("loading plugins configuration...")
	err := a.loadPlugins(c)
	if err != nil {
		a.logger.Errorf("error while loading plugins: %v", err)
		return err
	}

	a.logger.Info("initializing plugins...")
	err = a.initPlugins(c)
	if err != nil {
		a.logger.Errorf("error while initializing plugins: %v", err)
		return err
	}

	err = a.pipelineManger.StartAll()
	if err != nil {
		a.logger.Errorf("error while starting pipelines: %v", err)
		return err
	}

	// starting Mux
	a.start()

	a.logger.Info("starting output plugins...")
	err = a.startOutputPlugins()
	if err != nil {
		a.logger.Errorf("error while starting output plugins: %v", err)
		return err
	}

	a.logger.Info("starting processor plugins...")
	err = a.startProcessorPlugins()
	if err != nil {
		a.logger.Errorf("error while starting processor plugins: %v", err)
		return err
	}

	a.logger.Info("checking for any persisted async requests that need to be done...")
	err = a.pipelineManger.RecoverAsyncAll()
	if err != nil {
		a.logger.Errorf("error while applying persisted async requests: %v", err)
		return err
	}

	err = a.pipelineManger.Cleanup()
	if err != nil {
		a.logger.Errorf("error while applying persisted async requests: %v", err)
		return err
	}

	a.logger.Info("starting input plugins...")
	err = a.startInputPlugins()
	if err != nil {
		a.logger.Errorf("error while starting input plugins: %v", err)
		return err
	}

	a.logger.Info("successfully started all components")
	return nil
}

//stopComponents Stop components in graceful strategy
// 		1- Stop Input Components.
// 		2- Stop Pipelines.
// 		3- Stop Processor Components.
// 		4- Stop Output Components.
// As by definition each stop functionality in these components is graceful, this should guarantee graceful shutdown.
func (a *App) stopComponents() error {
	a.logger.Info("stopping all components gracefully...")

	///////////////////////////////////////

	a.logger.Info("stopping input plugins...")
	err := a.stopInputPlugins()
	if err != nil {
		a.logger.Errorf("failed to stop input plugins: %v", err)
		return err
	}
	a.logger.Info("stopped input plugins successfully.")

	///////////////////////////////////////

	a.logger.Info("stopping pipelines...")
	err = a.pipelineManger.StopAll()
	if err != nil {
		a.logger.Errorf("failed to stop pipelines: %v", err)
		return err
	}
	a.logger.Info("stopping pipelines successfully.")

	///////////////////////////////////////

	a.logger.Info("stopping processor plugins...")
	err = a.stopProcessorPlugins()
	if err != nil {
		a.logger.Errorf("failed to stop input plugins: %v", err)
		return err
	}
	a.logger.Info("stopping processor successfully.")

	///////////////////////////////////////

	a.logger.Info("stopping output plugins...")
	err = a.stopOutputPlugins()
	if err != nil {
		a.logger.Errorf("failed to stop output plugins: %v", err)
		return err
	}
	a.logger.Info("stopped output successfully.")

	///////////////////////////////////////

	a.logger.Info("stopped all components successfully")

	return nil
}
