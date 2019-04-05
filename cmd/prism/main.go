package main

import (
	"github.com/sherifabdlnaby/prism/app/config"
	input "github.com/sherifabdlnaby/prism/internal/input/dummy"
	output "github.com/sherifabdlnaby/prism/internal/output/disk"
	processor "github.com/sherifabdlnaby/prism/internal/processor/dummy"
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// USED FOR TESTING FOR NOW
func main() {

	// INIT LOGGER
	logConfig := zap.NewDevelopmentConfig()
	logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := logConfig.Build()
	defer logger.Sync()

	// READ CONFIG MAIN FILES
	inputConfig := config.InputsConfig{}
	err := config.Load("input_plugins.yaml", &inputConfig, true)
	if err != nil {
		panic(err)
	}
	processorConfig := config.ProcessorsConfig{}
	err = config.Load("processor_plugins.yaml", &processorConfig, true)
	if err != nil {
		panic(err)
	}
	outputConfig := config.OutputsConfig{}
	err = config.Load("output_plugins.yaml", &outputConfig, true)
	if err != nil {
		panic(err)
	}

	// output
	var outputDisk types.Output = &output.Disk{}
	outputLogger := logger.Named("output")
	outputPluginConfig := types.NewConfig(outputConfig.Outputs["dummyPlugin"].Config)

	// init & start output
	err = outputDisk.Init(*outputPluginConfig, *outputLogger.Named("disk"))
	if err != nil {
		panic(err)
	}
	err = outputDisk.Start()
	if err != nil {
		panic(err)
	}

	// processor
	processorDummy := processor.Dummy{}
	processorLogger := logger.Named("processor")
	processorPluginConfig := types.NewConfig(processorConfig.Processors["dummyPlugin"].Config)

	// init & start processor
	err = processorDummy.Init(*processorPluginConfig, *processorLogger.Named("dummy"))
	if err != nil {
		panic(err)
	}
	err = processorDummy.Start()
	if err != nil {
		panic(err)
	}

	// dummy
	var inputDummy types.Input = &input.Dummy{}
	inputLogger := logger.Named("dummy")
	inputPluginConfig := types.NewConfig(inputConfig.Inputs["dummyPlugin"].Config)

	// init & start dummy
	err = inputDummy.Init(*inputPluginConfig, *inputLogger.Named("dummy"))
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
