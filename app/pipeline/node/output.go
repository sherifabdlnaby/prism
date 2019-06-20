package node

import (
	"github.com/sherifabdlnaby/prism/app/registry/wrapper"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//output Wraps an output component
type output struct {
	output *wrapper.Output
	*Node
}

//NewOutput Construct a new Output Node
func NewOutput(out *wrapper.Output) *output {
	Node := &output{output: out}
	base := newBase(Node, out.Resource)
	Node.Node = base
	return Node
}

//job output job will send the transaction to output plugin and await its result.
func (n *output) job(t transaction.Transaction) {
	err := n.resource.Acquire(t.Context)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	responseChan := make(chan response.Response)

	n.output.TransactionChan <- transaction.Transaction{
		Payload:      t.Payload,
		Data:         t.Data,
		ResponseChan: responseChan,
		Context:      t.Context,
	}

	t.ResponseChan <- <-responseChan

	n.resource.Release()
}

func (n *output) jobStream(t transaction.Transaction) {
	n.job(t)
}
