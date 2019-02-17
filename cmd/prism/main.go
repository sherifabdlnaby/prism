package main

import (
	"github.com/sherifabdlnaby/prism/internal/input"
	"github.com/sherifabdlnaby/prism/internal/output"
	"github.com/sherifabdlnaby/prism/internal/processor"
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"time"
)

// USED FOR TESTING FOR NOW
func main() {

	// INIT LOGGER
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build()

	defer logger.Sync()

	// output
	var outputDummy types.Output = &output.Dummy{}
	outputLogger := logger.Named("output")
	outputConfig := types.Config{
		"filename": "output.jpg",
	}

	// init & start output
	_ = outputDummy.Init(outputConfig, *outputLogger.Named("dummy"))
	_ = outputDummy.Start()

	// processor
	processorDummy := processor.Dummy{}
	processorLogger := logger.Named("processor")

	// init & start processor
	_ = processorDummy.Init(nil, *processorLogger.Named("dummy"))
	_ = processorDummy.Start()

	// input
	var inputDummy types.Input = &input.Dummy{}
	inputLogger := logger.Named("input")

	// input outputConfig
	inputConfig := types.Config{
		"filename": "test.jpg",
	}
	// init & start input
	_ = inputDummy.Init(inputConfig, *inputLogger.Named("dummy"))
	_ = inputDummy.Start()

	outputNode := func(t types.Transaction) {
		outputDummy.TransactionChan() <- t
	}

	processorNode := func(t types.Transaction) {

		/// PROCESSING PART
		decoded, _ := processorDummy.Decode(t.EncodedPayload)
		decodedPayload, _ := processorDummy.Process(decoded)
		encoded, _ := processorDummy.Encode(decodedPayload)
		///

		responseChan := make(chan types.Response)

		go outputNode(types.Transaction{
			EncodedPayload: encoded,
			ResponseChan:   responseChan,
		})

		// forward response (no logic needed for now)
		t.ResponseChan <- <-responseChan

	}

	pipeline := func(st types.StreamableTransaction) {
		//Convert Streamable to non Streamable
		bytes, _ := ioutil.ReadAll(st)
		responseChan := make(chan types.Response)
		transaction := types.Transaction{
			EncodedPayload: types.EncodedPayload{
				Name:       "",
				ImageBytes: bytes,
				ImageData:  nil,
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

	time.Sleep(8 * time.Second)

	_ = inputDummy.Close(1 * time.Second)
	_ = processorDummy.Close(1 * time.Second)
	_ = outputDummy.Close(1 * time.Second)
}
