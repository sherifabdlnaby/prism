package main

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app"
	"github.com/sherifabdlnaby/prism/app/config"
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
	err = config.Load("input_plugins_BAK.yaml", &inputConfig, true)
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

	return config.Config{
		Inputs:     inputConfig,
		Processors: processorConfig,
		Outputs:    outputConfig,
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
	inputName := "dummy_input"

	// output
	outputLogger := config.Logger.Named("output")
	outputDisk := app.Registry[outputName].Constructor().(types.Output)
	outputPluginConfig := types.NewConfig(config.Outputs.Outputs[outputName].Config)

	// init & start output
	err = outputDisk.Init(*outputPluginConfig, *outputLogger.Named(outputName))
	if err != nil {
		panic(err)
	}
	err = outputDisk.Start()
	if err != nil {
		panic(err)
	}

	// processor
	processorLogger := config.Logger.Named("processor")
	processorDummy := app.Registry[processorName].Constructor().(types.Processor)
	processorPluginConfig := types.NewConfig(config.Processors.Processors[processorName].Config)

	// init & start processor
	err = processorDummy.Init(*processorPluginConfig, *processorLogger.Named(processorName))
	if err != nil {
		panic(err)
	}
	err = processorDummy.Start()
	if err != nil {
		panic(err)
	}

	// dummy
	inputLogger := config.Logger.Named("Input")
	inputDummy := app.Registry[inputName].Constructor().(types.Input)
	inputPluginConfig := types.NewConfig(config.Inputs.Inputs[inputName].Config)

	// init & start dummy
	err = inputDummy.Init(*inputPluginConfig, *inputLogger.Named(inputName))
	if err != nil {
		panic(err)
	}
	err = inputDummy.Start()
	if err != nil {
		panic(err)
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
