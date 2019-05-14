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
	err := a.loadPlugins(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.initPlugins(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.initPipelines(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.startPipelines(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	a.startMux()

	err = a.startOutputPlugins(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.startProcessorPlugins(c)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.startInputPlugins(c)
	if err != nil {
		a.logger.Error(err)
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

	a.logger.inputLogger.Info("stopping input plugins...")
	err := a.stopInputPlugins(c)
	if err != nil {
		a.logger.inputLogger.Errorw("failed to stop input plugins", "error", err.Error())
		return err
	}
	a.logger.inputLogger.Info("stopped input plugins successfully.")

	///////////////////////////////////////

	a.logger.pipelineLogger.Info("stopping pipelines...")
	err = a.stopPipelines(c)
	if err != nil {
		a.logger.pipelineLogger.Errorw("failed to stop pipelines", "error", err.Error())
		return err
	}
	a.logger.pipelineLogger.Info("stopping pipelines successfully.")

	///////////////////////////////////////

	err = a.stopProcessorPlugins(c)
	a.logger.processingLogger.Info("stopping processor plugins...")
	if err != nil {
		a.logger.processingLogger.Errorw("failed to stop input plugins", "error", err.Error())
		return err
	}
	a.logger.processingLogger.Info("stopping processor successfully.")

	///////////////////////////////////////

	err = a.stopOutputPlugins(c)
	a.logger.outputLogger.Info("stopping output plugins...")
	if err != nil {
		a.logger.outputLogger.Errorw("failed to stop output plugins", "error", err.Error())
		return err
	}
	a.logger.outputLogger.Info("stopped output successfully.")

	///////////////////////////////////////

	a.logger.Info("stopped all components successfully")

	return nil
}
