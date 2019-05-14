package app

import (
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline"
	"github.com/sherifabdlnaby/prism/app/registery"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//App is an self contained instance of Prism app.
type App struct {
	config    config.Config
	logger    logger
	registry  registery.Registry
	pipelines map[string]Pipeline
}

type Pipeline struct {
	pipeline.Pipeline
	TransactionChan chan transaction.Transaction
}

//NewApp Construct a new instance of Prism App using parsed config, instance still need to be initialized and started.
func NewApp(config config.Config) *App {

	app := &App{
		config:    config,
		logger:    *newLoggers(config),
		registry:  *registery.NewRegistry(),
		pipelines: make(map[string]Pipeline),
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

//startComponents startMux components
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

	a.logger.Info("initializing pipelines...")
	err = a.initPipelines(c)
	if err != nil {
		a.logger.Errorf("error while initializing pipelines: %v", err)
		return err
	}

	a.logger.Info("starting pipelines...")
	err = a.startPipelines(c)
	if err != nil {
		a.logger.Errorf("error while starting pipelines: %v", err)
		return err
	}

	a.startMux()

	a.logger.Info("starting output plugins...")
	err = a.startOutputPlugins(c)
	if err != nil {
		a.logger.Errorf("error while starting output plugins: %v", err)
		return err
	}

	a.logger.Info("starting processor plugins...")
	err = a.startProcessorPlugins(c)
	if err != nil {
		a.logger.Errorf("error while starting processor plugins: %v", err)
		return err
	}

	a.logger.Info("starting input plugins...")
	err = a.startInputPlugins(c)
	if err != nil {
		a.logger.Errorf("error while starting input plugins: %v", err)
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

	a.logger.Info("stopping input plugins...")
	err := a.stopInputPlugins(c)
	if err != nil {
		a.logger.Errorf("failed to stop input plugins: %v", err)
		return err
	}
	a.logger.Info("stopped input plugins successfully.")

	///////////////////////////////////////

	a.logger.Info("stopping pipelines...")
	err = a.stopPipelines(c)
	if err != nil {
		a.logger.Errorf("failed to stop pipelines: %v", err)
		return err
	}
	a.logger.Info("stopping pipelines successfully.")

	///////////////////////////////////////

	err = a.stopProcessorPlugins(c)
	a.logger.Info("stopping processor plugins...")
	if err != nil {
		a.logger.Errorf("failed to stop input plugins: %v", err)
		return err
	}
	a.logger.Info("stopping processor successfully.")

	///////////////////////////////////////

	err = a.stopOutputPlugins(c)
	a.logger.Info("stopping output plugins...")
	if err != nil {
		a.logger.Errorf("failed to stop output plugins: %v", err)
		return err
	}
	a.logger.Info("stopped output successfully.")

	///////////////////////////////////////

	a.logger.Info("stopped all components successfully")

	return nil
}
