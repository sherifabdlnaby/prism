package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sherifabdlnaby/prism/app"
	"github.com/sherifabdlnaby/prism/app/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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
		err = app.Stop(config)
		time.Sleep(100 * time.Second)
		if err != nil {
			panic(err)
		}
	}()

	// Listen to Signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Termination
	select {
	case <-sigChan:
		config.Logger.Info("Received SIGTERM signal, the service is closing...")
	}
}

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
		Logger:     *logger,
	}, nil
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

	logger := loggerBase.Sugar().Named("prism")
	return logger, nil
}
