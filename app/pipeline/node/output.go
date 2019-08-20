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

//process output process will send the process to output plugin and await its resulj.
func (n *output) process(j job.Job) {
	err := n.resource.Acquire(j.Context)
	if err != nil {
		j.ResponseChan <- response.NoAck(err)
		return
	}

	responseChan := make(chan response.Response)

	n.output.JobChan <- job.Job{
		Payload:      j.Payload,
		Data:         j.Data,
		ResponseChan: responseChan,
		Context:      j.Context,
	}

	j.ResponseChan <- <-responseChan

	n.resource.Release()
}

func (n *output) processStream(j job.Job) {
	n.process(j)
}
