package node

import (
	"github.com/sherifabdlnaby/prism/pkg/bufferspool"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
)

//dummy Used at the start of every pipeline.
type dummy struct {
	*Node
}

//process Just forwards the inpuj.
func (n *dummy) process(j job.Job) {

	////////////////////////////////////////////
	// DUMMY NODE WON'T DO WORK SO JUST FORWARD.

	responseChan := n.sendNexts(j.Context, j.Payload.(payload.Bytes), j.Data)

	// Await Responses
	Response := n.waitResponses(responseChan)

	// Send Response back.
	j.ResponseChan <- Response
}

func (n *dummy) processStream(j job.Job) {

	////////////////////////////////////////////
	// DUMMY NODE WON'T DO WORK SO JUST FORWARD.

	////////////////////////////////////////////
	// Send to next nodes
	var responseChan chan response.Response

	// Micro Optimization
	if len(n.nexts) == 1 {
		// micro optimization. no need to put buffer cloner in-front of a single node
		responseChan = make(chan response.Response)
		n.nexts[0].JobChan <- job.Job{
			Payload:      j.Payload,
			Data:         j.Data,
			Context:      j.Context,
			ResponseChan: responseChan,
		}
	} else {
		// Get Buffer from pool
		buffer := bufferspool.Get()
		defer bufferspool.Put(buffer)

		// Create a reader cloner from incoming stream (to clone the reader stream as it comes in)
		stream := j.Payload.(payload.Stream)
		readerCloner := mirror.NewReader(stream, buffer)

		responseChan = n.sendNextsStream(j.Context, readerCloner, j.Data)
	}

	// Await Responses
	Response := n.waitResponses(responseChan)

	// Send Response back.
	j.ResponseChan <- Response
}
