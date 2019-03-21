package main

import (
	input "github.com/sherifabdlnaby/prism/internal/input/dummy"
	output "github.com/sherifabdlnaby/prism/internal/output/amazon-s3"
	processor "github.com/sherifabdlnaby/prism/internal/processor/dummy"
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	var outputAmazonS3 types.Output = &output.S3{}
	outputLogger := logger.Named("output")
	outputConfig := types.Config{
		"filepath":  "/home/output.jpg",
		"s3_region": "us-east-2",
		"s3_bucket": "prism.test",
	}

	// init & start output
	_ = outputAmazonS3.Init(outputConfig, *outputLogger.Named("s3"))
	_ = outputAmazonS3.Start()

	// processor
	processorDummy := processor.Dummy{}
	processorLogger := logger.Named("processor")

	// init & start processor
	_ = processorDummy.Init(nil, *processorLogger.Named("dummy"))
	_ = processorDummy.Start()

	// dummy
	var inputDummy types.Input = &input.Dummy{}
	inputLogger := logger.Named("dummy")

	// dummy outputConfig
	inputConfig := types.Config{
		"filename": "test.jpg",
	}
	// init & start dummy
	_ = inputDummy.Init(inputConfig, *inputLogger.Named("dummy"))
	_ = inputDummy.Start()

	outputNode := func(t types.Transaction) {
		outputAmazonS3.TransactionChan() <- t
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
	_ = outputAmazonS3.Close(1 * time.Second)

	time.Sleep(1 * time.Second)
}