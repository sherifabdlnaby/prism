package node

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//Output Wraps an output component
type Output struct {
	component.Output
	receiveChan <-chan transaction.Transaction
	Next        []Next
	Resource    Resource
}

//startMux startMux receiving transactions
func (n *Output) Start() error {
	go func() {
		for value := range n.receiveChan {
			go n.job(value)
		}
	}()
	return nil
}

func (n *Output) Stop() error {
	for _, value := range n.Next {
		close(value.TransactionChan)
	}
	return nil
}

//job Output job will send the transaction to output plugin and await its result.
func (n *Output) job(t transaction.Transaction) {
	err := n.Resource.Acquire(t.Context, 1)
	if err != nil {
		t.ResponseChan <- response.NoAck(err)
		return
	}

	responseChan := make(chan response.Response, len(n.Next))

	n.TransactionChan() <- transaction.Transaction{
		Payload:      t.Payload,
		ImageData:    t.ImageData,
		ResponseChan: responseChan,
		Context:      t.Context,
	}

	t.ResponseChan <- <-responseChan

	n.Resource.Release(1)
}

func (n *Output) SetTransactionChan(tc <-chan transaction.Transaction) {
	n.receiveChan = tc
}