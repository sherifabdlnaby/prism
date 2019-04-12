package main

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/manager"
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

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

	outputName := "disk"
	processorName := "dummy_processor"
	inputName := "http_server"

	err = app.LoadPlugins(config)

	if err != nil {
		config.Logger.Panic(err)
	}

	err = app.InitPlugins(config)
	if err != nil {
		config.Logger.Panic(err)
	}

	// output
	outputDisk := manager.Get(outputName).(types.Output)
	err = outputDisk.Start()
	if err != nil {
		config.Logger.Panic(err)
	}

	// processor
	processorDummy := manager.Get(processorName).(types.Processor)
	err = processorDummy.Start()
	if err != nil {
		config.Logger.Panic(err)
	}

	// dummy
	inputDummy := manager.Get(inputName).(types.Input)
	err = inputDummy.Start()
	if err != nil {
		config.Logger.Panic(err)
	}

	outputNode := func(t types.Transaction) {
		outputDisk.TransactionChan() <- t
	}

	processorNode := func(t types.Transaction) {

		/// PROCESSING PART
		decoded, _ := processorDummy.Decode(t.Payload)
		decodedPayload, _ := processorDummy.Process(decoded)
		encoded, _ := processorDummy.Encode(decodedPayload)
		///

		responseChan := make(chan types.Response)

		go outputNode(types.Transaction{
			Payload:      encoded,
			ResponseChan: responseChan,
		})

		// forward response (no logic needed for now)
		t.ResponseChan <- <-responseChan

	}

	pipeline := func(st types.Transaction) {
		responseChan := make(chan types.Response)
		transaction := types.Transaction{
			Payload: types.Payload{
				Name:      "",
				Reader:    st,
				ImageData: nil,
			},
			ResponseChan: responseChan,
		}

		go processorNode(transaction)

		// wait for response from processor
		st.ResponseChan <- <-responseChan
	}

	/// Harvest Input
	go func() {
		inputChan := inputDummy.TransactionChan()
		for streamableInput := range inputChan {
			go pipeline(streamableInput)
		}
	}()

	time.Sleep(2 * time.Second)

	_ = inputDummy.Close(1 * time.Second)
	_ = processorDummy.Close(1 * time.Second)
	_ = outputDisk.Close(1 * time.Second)

	time.Sleep(1 * time.Second)
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
