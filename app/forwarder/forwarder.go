package forwarder

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

type Forwarder struct {
	inputChans []<-chan transaction.InputTransaction
	pipelines  map[string]chan<- transaction.Transaction
}

func NewForwarder(inputChans []<-chan transaction.InputTransaction, pipelines map[string]chan<- transaction.Transaction) *Forwarder {
	return &Forwarder{inputChans: inputChans, pipelines: pipelines}
}

//Start starts the Forwarder that forwards the transactions from input to pipelines based on PipelineTag in transaction.
func (m *Forwarder) Start() {
	for _, value := range m.inputChans {
		go m.forwardInputToPipeline(value)
	}
}

func (m *Forwarder) forwardInputToPipeline(input <-chan transaction.InputTransaction) {

	for in := range input {

		//get pipeline tag
		if in.PipelineTag == "" {
			in.ResponseChan <- response.Error(fmt.Errorf("no pipeline is defined for in"))
			continue
		}

		_, ok := m.pipelines[in.PipelineTag]
		if !ok {
			in.ResponseChan <- response.Error(fmt.Errorf("pipeline [%s] is not defined", in.PipelineTag))
			continue
		}

		// Add defaults to in Image Data
		applyDefaultFields(in.Data)

		// Forward
		m.pipelines[in.PipelineTag] <- in.Transaction
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
