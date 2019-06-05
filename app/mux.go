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
		applyDefaultFields(Tchan.Data)

		a.pipelines[Tchan.PipelineTag].TransactionChan <- Tchan.Transaction
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
