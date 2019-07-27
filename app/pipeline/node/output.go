package node

import (
	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/response"
)

//output Wraps an output core
type output struct {
	output *component.Output
	*Node
}

//process output process will send the process to output plugin and await its result.
func (n *output) process(t job.Job) {
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

func (n *output) processStream(t job.Job) {
	n.process(t)
}
