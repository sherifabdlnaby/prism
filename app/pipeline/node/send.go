package node

import (
	"context"

	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
)

func (n *Node) sendNextsStream(ctx context.Context, writerCloner mirror.Cloner, data payload.Data) chan response.Response {
	responseChan := make(chan response.Response, len(n.nexts))

	for _, next := range n.nexts {
		// Copy new map
		newData := make(payload.Data, len(data))
		for key := range data {
			newData[key] = data[key]
		}

		next.JobChan <- job.Job{
			Payload:      writerCloner.Clone(),
			Data:         newData,
			Context:      ctx,
			ResponseChan: responseChan,
		}
	}
	return responseChan
}

func (n *Node) sendNexts(ctx context.Context, output payload.Bytes, data payload.Data) chan response.Response {
	responseChan := make(chan response.Response, len(n.nexts))

	for _, next := range n.nexts {
		// Copy new map
		newData := make(payload.Data, len(data))
		for key := range data {
			newData[key] = data[key]
		}

		next.JobChan <- job.Job{
			Payload:      output,
			Data:         newData,
			Context:      ctx,
			ResponseChan: responseChan,
		}
	}
	return responseChan
}

func (n *Node) waitResponses(responseChan chan response.Response) response.Response {
	////////////////////////////////////////////
	// receive from next nodes
	count, total := 0, len(n.nexts)
	Response := response.Response{}

	for ; count < total; count++ {
		Response = <-responseChan
		if !Response.Ack {
			break
		}
	}

	return Response
}
