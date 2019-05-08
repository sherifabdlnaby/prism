package mux

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sherifabdlnaby/prism/app/pipeline"
	"github.com/sherifabdlnaby/prism/app/registery/wrapper"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//Mux used to forward transactions coming from input plugins to it's pipelines based on transaction.PipelineTag, Mux
//also add default values to the transaction.ImageData in each transaction.
type Mux struct {
	Pipelines map[string]*pipeline.Pipeline
	Inputs    map[string]*wrapper.Input
}

//Start starts the mux that forwards the transactions from input to pipelines based on PipelineTag in transaction.
func (m *Mux) Start() {
	for _, value := range m.Inputs {
		go m.forwardPerInput(value)
	}
}

func (m *Mux) forwardPerInput(input *wrapper.Input) {
	for Tchan := range input.TransactionChan() {

		_, ok := m.Pipelines[Tchan.PipelineTag]
		if !ok {
			Tchan.ResponseChan <- response.Error(
				fmt.Errorf("pipeline [%s] is not defined", Tchan.PipelineTag),
			)
			continue
		}

		// Add defaults to transaction Image Data
		applyDefaultFields(Tchan.ImageData)

		m.Pipelines[Tchan.PipelineTag].ReceiveChan <- Tchan.Transaction
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
