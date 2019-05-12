package app

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sherifabdlnaby/prism/app/registery/wrapper"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//startMux starts the mux that forwards the transactions from input to pipelines based on PipelineTag in transaction.
func (a *App) startMux() {
	for _, value := range a.registry.InputPlugins {
		go a.forwardPerInput(value)
	}
}

func (a *App) forwardPerInput(input *wrapper.Input) {
	for Tchan := range input.TransactionChan() {

		_, ok := a.pipelines[Tchan.PipelineTag]
		if !ok {
			Tchan.ResponseChan <- response.Error(
				fmt.Errorf("pipeline [%s] is not defined", Tchan.PipelineTag),
			)
			continue
		}

		// Add defaults to transaction Image Data
		applyDefaultFields(Tchan.ImageData)

		a.pipelines[Tchan.PipelineTag].ReceiveChan <- Tchan.Transaction
	}
}

func applyDefaultFields(d transaction.ImageData) {
	id := uuid.New()
	epoch := time.Now().Unix()
	addDefaultValueToMap(d, "_id", id.String())
	addDefaultValueToMap(d, "_timestamp", epoch)
}

func addDefaultValueToMap(data transaction.ImageData, key string, val interface{}) {
	_, ok := data[key]
	if ok {
		return
	}
	data[key] = val
	return
}
