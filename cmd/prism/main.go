package main

import (
	"fmt"
	"time"

	"github.com/sherifabdlnaby/prism/app"
	"github.com/sherifabdlnaby/prism/app/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// PARSE STUFF
func bootstrap() (config.Config, error) {
	// READ CONFIG MAIN FILES
	appConfig := config.AppConfig{}
	err := config.Load("prism.yaml", &appConfig, true)
	if err != nil {
		return config.Config{}, err
	}

	// INIT LOGGER
	logger, err := bootLogger(appConfig)
	if err != nil {
		return config.Config{}, err
	}

	// READ CONFIG MAIN FILES
	inputConfig := config.InputsConfig{}
	err = config.Load("input_plugins.yaml", &inputConfig, true)
	if err != nil {
		return config.Config{}, err
	}

	processorConfig := config.ProcessorsConfig{}
	err = config.Load("processor_plugins.yaml", &processorConfig, true)
	if err != nil {
		return config.Config{}, err
	}
	outputConfig := config.OutputsConfig{}
	err = config.Load("output_plugins.yaml", &outputConfig, true)
	if err != nil {
		return config.Config{}, err
	}

	pipelineConfig := config.PipelinesConfig{}
	err = config.Load("pipeline.yaml", &pipelineConfig, true)
	if err != nil {
		return config.Config{}, err
	}

	return config.Config{
		App:        appConfig,
		Inputs:     inputConfig,
		Processors: processorConfig,
		Outputs:    outputConfig,
		Pipeline:   pipelineConfig,
		Logger:     logger,
	}, nil
}

// USED FOR TESTING FOR NOW
func main() {

	config, err := bootstrap()

	if err != nil {
		panic(err)
	}

	app := app.NewApp(config)

	err = app.InitializeComponents(config)
	err = app.InitializePipelines(config)
	err = app.StartPipelines(config)
	err = app.StartMux(config)
	err = app.StartComponents(config)

	time.Sleep(5 * time.Second)

	err = app.StopComponents(config)
	if err != nil {
		panic(err)
	}
}

func bootLogger(appConfig config.AppConfig) (*zap.SugaredLogger, error) {
	var logConfig zap.Config
	if appConfig.Logger == "dev" {
		logConfig = zap.NewDevelopmentConfig()
		logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else if appConfig.Logger == "prod" {
		logConfig = zap.NewProductionConfig()
	} else {
		return nil, fmt.Errorf("logger config can be either \"dev\" or \"prod\"")
	}

	loggerBase, err := logConfig.Build()

	if err != nil {
		return nil, err
	}

	logger := loggerBase.Sugar()
	return logger, nil
}
