package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline"
	componentConfig "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

// loadPlugins Load all plugins in Config
func (a *App) loadPlugins(c config.Config) error {

	// Load Input Plugins
	for name, plugin := range c.Inputs.Inputs {
		err := a.registry.LoadInput(name, *plugin, a.logger.inputLogger)
		if err != nil {
			return err
		}
	}

	// Load Processor Plugins
	for name, plugin := range c.Processors.Processors {
		err := a.registry.LoadProcessor(name, *plugin, a.logger.processingLogger)
		if err != nil {
			return err
		}
	}

	// Load Output Plugins
	for name, plugin := range c.Outputs.Outputs {
		err := a.registry.LoadOutput(name, *plugin, a.logger.outputLogger)
		if err != nil {
			return err
		}
	}

	return nil
}

// initPlugins Init all plugins in Config by calling their Init() function
func (a *App) initPlugins(c config.Config) error {

	// Init Input Plugins
	for name, input := range c.Inputs.Inputs {
		plugin, _ := a.registry.GetInput(name)
		pluginConfig := *componentConfig.NewConfig(input.Config)
		err := plugin.Init(pluginConfig, *a.logger.inputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initialize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Processor Plugins
	for name, processor := range c.Processors.Processors {
		plugin, _ := a.registry.GetProcessor(name)
		pluginConfig := *componentConfig.NewConfig(processor.Config)
		err := plugin.Init(pluginConfig, *a.logger.processingLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initialize plugin [%s]: %s", name, err.Error())
		}
	}

	// Load Output Plugins
	for name, output := range c.Outputs.Outputs {
		plugin, _ := a.registry.GetOutput(name)
		pluginConfig := *componentConfig.NewConfig(output.Config)
		err := plugin.Init(pluginConfig, *a.logger.outputLogger.Named(name))
		if err != nil {
			return fmt.Errorf("failed to initialize plugin [%s]: %s", name, err.Error())
		}
	}

	return nil
}

// startInputPlugins start all input plugins in Config by calling their start() function
func (a *App) startInputPlugins() error {

	for name, value := range a.registry.Inputs {
		err := value.Start()
		if err != nil {
			a.logger.inputLogger.Errorf("failed to start input plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// startProcessorPlugins start all processor plugins in Config by calling their start() function
func (a *App) startProcessorPlugins() error {

	for name, value := range a.registry.Processors {
		err := value.Start()
		if err != nil {
			a.logger.processingLogger.Errorf("failed to start processor plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// startOutputPlugins start all output plugins in Config by calling their start() function
func (a *App) startOutputPlugins() error {

	for name, value := range a.registry.Outputs {
		err := value.Start()
		if err != nil {
			a.logger.outputLogger.Errorf("failed to start output plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// initPipelines Initialize and build all configured pipelines
func (a *App) initPipelines(c config.Config) error {

	dataDir := config.EnvPrismDataDir.Lookup()

	// open pipeline DB
	err := os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll(config.EnvPrismTmpDir.Lookup(), os.ModePerm)
	if err != nil {
		return err
	}

	db, err := bolt.Open(dataDir+"/async.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return fmt.Errorf("error while opening persistence DB file %s", err.Error())
	}

	for key, pipConfig := range c.Pipeline.Pipelines {

		// check if pipeline already exists
		_, ok := a.pipelines[key]
		if ok {
			return fmt.Errorf("pipeline with name [%s] already declared", key)
		}

		pip, err := pipeline.NewPipeline(key, db, *pipConfig, a.registry, *a.logger.processingLogger.Named(key), c.Pipeline.Hash)

		if err != nil {
			return fmt.Errorf("error occurred when constructing pipeline [%s]: %s", key, err.Error())
		}

		tc := make(chan transaction.Transaction)
		pip.SetTransactionChan(tc)

		a.pipelines[key] = pipelineWrapper{
			Pipeline:        pip,
			TransactionChan: tc,
		}
	}

	return nil
}

// startPipelines start all pipelines and start accepting input
func (a *App) startPipelines() error {

	for _, value := range a.pipelines {
		err := value.Start()
		if err != nil {
			value.Logger.Error(err.Error())
			return err
		}
	}

	return nil
}

// stopPipelines Stop pipelines by calling their Stop() function, any request to these pipelines will return error.
func (a *App) stopPipelines() error {

	for _, value := range a.pipelines {
		// close receiving chan
		close(value.TransactionChan)

		// stop pipeline
		err := value.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

// stopPipelines Stop pipelines by calling their Stop() function, any request to these pipelines will return error.
func (a *App) applyPersistedAsyncRequests() error {
	tmpPath := config.EnvPrismTmpDir.Lookup()

	//save current files
	files, err := ioutil.ReadDir(tmpPath)
	if err != nil {
		return err
	}

	go func() {
		for _, value := range a.pipelines {
			err := value.ApplyPersistedAsyncRequests()
			if err != nil {
				a.logger.Errorw("error while applying lost async requests", "error", err.Error())
			}
		}

		// Here we remove any files we read before applying the requests,
		// this for the narrow possibility that a system crash happened after removing it from DB but before deleting it from the file,
		// as the DB is the source of truth, any remaining file after applying everything shall be removed.
		// Need to check if they're not removed After we read them as applying itself could have removed its own files.
		// Bottom line this function remove any image that has no entry in the DB.
		for _, file := range files {
			if !file.IsDir() {
				// get abs Path
				filePath := tmpPath + "/" + file.Name()
				filePath, _ = filepath.Abs(filePath)
				if _, err := os.Stat(filePath); !os.IsNotExist(err) {
					err := os.Remove(filePath)
					if err != nil {
						a.logger.Errorw("error while cleaning up tmp images", "error", err.Error())
					}
				}
			}
		}

	}()

	return nil
}

// stopInputPlugins Stop all input plugins in Config by calling their Stop() function
func (a *App) stopInputPlugins() error {

	for name, value := range a.registry.Inputs {
		err := value.Close()
		if err != nil {
			a.logger.inputLogger.Errorf("failed to stop input plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// stopProcessorPlugins Stop all processor plugins in Config by calling their Stop() function
func (a *App) stopProcessorPlugins() error {

	for name, value := range a.registry.Processors {
		err := value.Close()
		if err != nil {
			a.logger.processingLogger.Errorf("failed to stop processor plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}

// stopOutputPlugins Stop all output plugins in Config by calling their Stop() function
func (a *App) stopOutputPlugins() error {

	for name, value := range a.registry.Outputs {
		// close its transaction chan (stop sending txns to it)
		close(value.TransactionChan)

		// close plugin
		err := value.Close()

		if err != nil {
			a.logger.outputLogger.Errorf("failed to stop output plugin [%s]: %v", name, err)
			return err
		}
	}

	return nil
}
