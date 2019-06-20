package node

import (
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//Node A Node wraps components and manage receiving transactions and forwarding transactions to next nodes.
type Node interface {
	// Start this Node and all its next nodes to start receiving transactions
	Start() error

	//Stop Stop this Node and stop all its next nodes.
	Stop() error

	//SetTransactionChan Set the transaction chan Node will use to receive input
	SetTransactionChan(<-chan transaction.Transaction)

	//SetAsync Set if this Node is sync/async
	SetAsync(bool)

	//SetNexts Set this Node's next nodes.
	SetNexts([]Next)
}

//TODO rename
type component interface {
	job(t transaction.Transaction)
	jobStream(t transaction.Transaction)
}

//Next Wraps the next Node plus the channel used to communicate with this Node to send input transactions.
type Next struct {
	Node
	TransactionChan chan transaction.Transaction
}

//NewNext Create a new Next Node with the supplied Node.
func NewNext(node Node) *Next {
	return &Next{
		Node:            node,
		TransactionChan: make(chan transaction.Transaction),
	}
}
