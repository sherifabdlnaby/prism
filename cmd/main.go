package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sherifabdlnaby/prism/app"
	"github.com/sherifabdlnaby/prism/app/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Listen to Signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Parse configuration from yaml files
	config, err := bootstrap()
	if err != nil {
		panic(err)
	}

	// Create new app instance
	app := app.NewApp(config)

	// start app
	err = app.Start(config)
	if err != nil {
		panic(err)
	}

	// Defer Closing the app.
	defer func() {
		err = app.Stop()
		if err != nil {
			panic(err)
		}
	}()

	// Termination
	sig := <-signalChan

	config.Logger.Infof("Received %s signal, the service is closing...", sig.String())

}

// PARSE STUFF
func bootstrap() (config.Config, error) {

	// Setup ENVIRONMENT
	environment := config.PRISM_ENV.Lookup()
	if environment != "prod" && environment != "dev" {
		panic(fmt.Sprintf("Environment = \"%s\" (set by %s) can only be either \"prod\" or \"dev\" (default: prod)", environment, config.PRISM_ENV.Name()))
	}

	// Print logo (Yes I love this)
	if environment == "dev" {
		printLogo()
	}

	// Initialize Root Logger
	logger, err := bootLogger(environment)
	if err != nil {
		return config.Config{}, err
	}

	// Log environment
	logger.Infof("logger initialized for %s environment", environment)

	// Load .env environment variable if in environment is dev
	if environment == "dev" {
		err = godotenv.Load()
		if err == nil {
			logger.Info("loaded .env into environment variables")
		}
	}

	// GET YAML FILE DIRECTORY
	configDir := config.PRISM_CONFIG_DIR.Lookup()

	// use full path
	configDir, err = filepath.Abs(configDir)
	if err != nil {
		return config.Config{}, err
	}

	// Log environment
	logger.Infof("loading config files from %s", configDir)

	appConfigPath := configDir + "/prism.yaml"
	inputConfigPath := configDir + "/inputs.yaml"
	outputConfigPath := configDir + "/outputs.yaml"
	processorConfigPath := configDir + "/processors.yaml"
	pipelineConfigPath := configDir + "/pipelines.yaml"

	// READ CONFIG MAIN FILES
	appConfig := config.AppConfig{}
	err = config.Load(appConfigPath, &appConfig, true)
	if err != nil {
		return config.Config{}, err
	}

	// READ CONFIG MAIN FILES
	inputConfig := config.InputsConfig{}
	err = config.Load(inputConfigPath, &inputConfig, true)
	if err != nil {
		return config.Config{}, err
	}

	processorConfig := config.ProcessorsConfig{}
	err = config.Load(processorConfigPath, &processorConfig, true)
	if err != nil {
		return config.Config{}, err
	}
	outputConfig := config.OutputsConfig{}
	err = config.Load(outputConfigPath, &outputConfig, true)
	if err != nil {
		return config.Config{}, err
	}

	pipelineConfig := config.PipelinesConfig{}
	err = config.Load(pipelineConfigPath, &pipelineConfig, true)
	if err != nil {
		return config.Config{}, err
	}

	return config.Config{
		App:        appConfig,
		Inputs:     inputConfig,
		Processors: processorConfig,
		Outputs:    outputConfig,
		Pipeline:   pipelineConfig,
		Logger:     *logger,
	}, nil
}

func bootLogger(env string) (*zap.SugaredLogger, error) {
	var logConfig zap.Config
	if env == "dev" {
		logConfig = zap.NewDevelopmentConfig()
		logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else if env == "prod" {
		logConfig = zap.NewProductionConfig()
	}

	loggerBase, err := logConfig.Build()

	if err != nil {
		return nil, err
	}

	logger := loggerBase.Sugar().Named("prism")
	return logger, nil
}

func printLogo() {
	fmt.Print(`

 ________    ________      ___      ________       _____ ______      
|\   __  \  |\   __  \    |\  \    |\   ____\     |\   _ \  _   \    
\ \  \|\  \ \ \  \|\  \   \ \  \   \ \  \___|_    \ \  \\\__\ \  \   
 \ \   ____\ \ \   _  _\   \ \  \   \ \_____  \    \ \  \\|__| \  \  
  \ \  \___|  \ \  \\  \|   \ \  \   \|____|\  \    \ \  \    \ \  \ 
   \ \__\      \ \__\\ _\    \ \__\    ____\_\  \    \ \__\    \ \__\
    \|__|       \|__|\|__|    \|__|   |\_________\    \|__|     \|__|
                                      \|_________|                   

`)
}
