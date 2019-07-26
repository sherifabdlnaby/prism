package node

import (
	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/response"
)

//output Wraps an output Type
type output struct {
	output *component.Output
	*Node
}

//job output job will send the job to output plugin and await its result.
func (n *output) job(t job.Job) {
	err := n.resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	responseChan := make(chan response.Response)

	n.output.JobChan <- job.Job{
		Payload:      t.Payload,
		Data:         t.Data,
		ResponseChan: responseChan,
		Context:      t.Context,
	}

	t.ResponseChan <- <-responseChan

	n.resource.Release()
}

func (n *output) jobStream(t job.Job) {
	n.job(t)
}
