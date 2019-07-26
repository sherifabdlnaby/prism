package forwarder

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
)

type Forwarder struct {
	inputChans []<-chan job.Input
	pipelines  map[string]chan<- job.Job
}

func NewForwarder(inputChans []<-chan job.Input, pipelines map[string]chan<- job.Job) *Forwarder {
	return &Forwarder{inputChans: inputChans, pipelines: pipelines}
}

//Start starts the Forwarder that forwards the jobs from input to pipelines based on PipelineTag in job.
func (m *Forwarder) Start() {
	for _, value := range m.inputChans {
		go m.forwardInputToPipeline(value)
	}
}

func (m *Forwarder) forwardInputToPipeline(input <-chan job.Input) {

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
		m.pipelines[in.PipelineTag] <- in.Job
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
