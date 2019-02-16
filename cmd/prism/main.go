package main

import (
	"github.com/sherifabdlnaby/prism/lib/pkg/input"
	"github.com/sherifabdlnaby/prism/lib/pkg/output"
	"github.com/sherifabdlnaby/prism/lib/pkg/processor"
	"github.com/sherifabdlnaby/prism/lib/pkg/types"
	"io/ioutil"
	"time"
)

// USED FOR TESTING FOR NOW
func main() {

	// output
	outputDummy := output.Dummy{}

	// output outputConfig
	outputConfig := types.Config{
		"filename": "output.jpg",
	}

	// init & start output
	_ = outputDummy.Init(outputConfig)
	_ = outputDummy.Start()

	// processor
	processorDummy := processor.Dummy{}

	// init & start processor
	_ = processorDummy.Init(nil)
	_ = processorDummy.Start()

	// input
	var inputDummy types.Input = &input.Dummy{}

	// input outputConfig
	inputConfig := types.Config{
		"filename": "test.jpg",
	}
	// init & start input
	_ = inputDummy.Init(inputConfig)
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

	_ = outputDummy.Close(1 * time.Second)
	_ = processorDummy.Close(1 * time.Second)
	_ = inputDummy.Close(1 * time.Second)
}
