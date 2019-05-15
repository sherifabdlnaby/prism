package node

import (
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//output Wraps an output component
type output struct {
	output component.Output
	*base
}

//NewOutput Construct a new Output Node
func NewOutput(out component.Output, r resource.Resource) Node {
	Node := &output{output: out}
	base := newBase(Node, r)
	Node.base = base
	return Node
}

//job output job will send the transaction to output plugin and await its result.
func (n *output) job(t transaction.Transaction) {
	err := n.resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	// if nodeType is set async, send response now, and navigate actual response to asyncResponses which should handle async responses
	if n.async {
		t.ResponseChan <- response.Ack()

		// used so that stop() wait for async responses to finish. (may be improved later)
		n.wg.Add(1)

		t.ResponseChan = n.asyncResponses
	}

	responseChan := make(chan response.Response)

	n.output.TransactionChan() <- transaction.Transaction{
		Payload:      t.Payload,
		ImageData:    t.ImageData,
		ResponseChan: responseChan,
		Context:      t.Context,
	}

	t.ResponseChan <- <-responseChan

	n.resource.Release()
}
