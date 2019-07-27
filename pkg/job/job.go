package job

import (
	"context"
	"os"

	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
)

// Job represent a job containing a streamable payload (the message) and a response channel,
// which is used to indicate whether the payload was successfully processed and propagated to the next destinations.
type Job struct {
	// Payload is the message payload of this job.
	// is either a payload.Stream OR payload.Bytes.
	Payload payload.Payload

	// Data is the message data of this job.
	Data payload.Data

	// Context of the job
	Context context.Context

	// ResponseChan should receive a response at the end of a job,
	// The response itself indicates whether the payload was successfully processed and propagated
	// to the next destinations.
	ResponseChan chan<- response.Response
}

// Input represent a job containing a streamable payload, Data, PipelineTag, and a response channel,
// response indicate whether the payload was successfully processed and propagated to the next destinations.
// PipelineTag indicate to which pipeline should this job be forwarded to.
type Input struct {
	Job
	PipelineTag string
}

// Async to be persisted in local DB
type Async struct {
	ID, Node, Filepath string
	Data               payload.Data
	Job                Job                      `json:"-"`
	JobResponseChan    <-chan response.Response `json:"-"`
}

func (a *Async) Load(Payload payload.Payload) error {
	var err error
	newPayload := Payload
	responseChan := make(chan response.Response, 1)

	if Payload == nil {
		// open tmp file
		newPayload, err = os.Open(a.Filepath)
		if err != nil {
			return err
		}
	}

	a.Job = Job{
		Payload:      newPayload,
		Data:         a.Data,
		Context:      context.Background(),
		ResponseChan: responseChan,
	}

	a.JobResponseChan = responseChan

	return nil
}
