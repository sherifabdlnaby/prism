package app

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sherifabdlnaby/prism/app/registry/wrapper"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
)

//start starts the mux that forwards the transactions from input to pipelines based on PipelineTag in transaction.
func (a *App) start() {
	for _, value := range a.registry.Inputs {
		go a.forwardInputToPipeline(value)
	}
}

func (a *App) forwardInputToPipeline(input *wrapper.Input) {
	for in := range input.InputTransactionChan() {

		//get pipeline tag
		if in.PipelineTag == "" {
			in.ResponseChan <- response.Error(fmt.Errorf("no pipeline is defined for in"))
			continue
		}

		_, ok := a.pipelines[in.PipelineTag]
		if !ok {
			in.ResponseChan <- response.Error(fmt.Errorf("pipeline [%s] is not defined", in.PipelineTag))
			continue
		}

		// Add defaults to in Image Data
		applyDefaultFields(in.Data)

		// Forward
		a.pipelines[in.PipelineTag].TransactionChan <- in.Transaction
	}
}

func applyDefaultFields(d payload.Data) {
	id := uuid.New()
	epoch := time.Now().Unix()
	addDefaultValueToMap(d, "_id", id.String())
	addDefaultValueToMap(d, "_timestamp", epoch)
}

func addDefaultValueToMap(data payload.Data, key string, val interface{}) {
	_, ok := data[key]
	if ok {
		return
	}
	data[key] = val
}
