package mux

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/pipeline"
	"github.com/sherifabdlnaby/prism/app/registery/wrapper"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//Mux used to forward transactions coming from input plugins to it's pipelines based on transaction.PipelineTag
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
			Tchan.ResponseChan <- transaction.ResponseError(
				fmt.Errorf("pipeline [%s] is not defined", Tchan.PipelineTag),
			)
			continue
		}

		m.Pipelines[Tchan.PipelineTag].RecieveChan <- Tchan.Transaction
	}
}
