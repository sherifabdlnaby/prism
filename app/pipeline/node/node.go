package node

import (
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//Node A nodeType wraps components and manage receiving transactions and forwarding transactions to next nodes.
type Node interface {
	// Start this nodeType and all its next nodes to start receiving transactions
	Start() error

	//Stop Stop this Node and stop all its next nodes.
	Stop() error

	//SetTransactionChan Set the transaction chan nodeType will use to receive input
	SetTransactionChan(<-chan transaction.Transaction)

	//SetTransactionChan Set the transaction chan nodeType will use to receive input
	SetStreamTransactionChan(<-chan transaction.Streamable)

	//SetAsync Set if this nodeType is sync/async
	SetAsync(bool)

	//SetNexts Set this nodeType's next nodes.
	SetNexts([]Next)
}

//TODO rename
type nodeType interface {
	job(t transaction.Transaction)
	jobStream(t transaction.Streamable)
}

//Next Wraps the next nodeType plus the channel used to communicate with this nodeType to send input transactions.
type Next struct {
	Node
	TransactionChan       chan transaction.Transaction
	StreamTransactionChan chan transaction.Streamable
}

//NewNext Create a new Next nodeType with the supplied Node.
func NewNext(Node Node) *Next {
	return &Next{
		Node:                  Node,
		TransactionChan:       make(chan transaction.Transaction),
		StreamTransactionChan: make(chan transaction.Streamable),
	}
}
