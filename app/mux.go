package app

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sherifabdlnaby/prism/app/registery/wrapper"
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
	for transaction := range input.TransactionChan() {

		//get pipeline tag
		tag, err := a.getValidPipelineTag(transaction.Data)

		if err != nil {
			transaction.ResponseChan <- response.Error(err)
			continue
		}

		// Add defaults to transaction Image Data
		applyDefaultFields(transaction.Data)

		// Forward
		a.pipelines[tag].TransactionChan <- transaction
	}
}

func (a *App) getValidPipelineTag(data payload.Data) (string, error) {
	tag, ok := data["_pipeline"]
	if !ok {
		return "", fmt.Errorf("no pipeline is defined for transaction")
	}

	// validate tag is string
	tagString, ok := tag.(string)
	if !ok {
		return "", fmt.Errorf("pipeline tag is not a string")
	}

	_, ok = a.pipelines[tagString]
	if !ok {
		return "", fmt.Errorf("pipeline [%s] is not defined", tagString)
	}

	return tagString, nil
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
