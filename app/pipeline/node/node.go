package node

import "github.com/sherifabdlnaby/prism/pkg/transaction"

// TODO refactor nodes to make code more reusable

// TODO enhance interfaces, avoid buffering reader if it's only one node in next.

//Node A node wraps components and manage receiving transactions and forwarding transactions to next nodes.
type Node interface {
	// Start this node and all its next nodes to start receiving transactions
	Start() error

	//Stop Stop this Node and stop all its next nodes.
	Stop() error

	//SetTransactionChan Set the transaction chan node will use to receive input
	SetTransactionChan(<-chan transaction.Transaction)

	//SetAsync Set if this node is sync/async
	SetAsync(bool)

	//SetNexts Set this node's next nodes.
	SetNexts([]Next)
}

//Next Wraps the next node plus the channel used to communicate with this node to send input transactions.
type Next struct {
	Node
	TransactionChan chan transaction.Transaction
}

//NewNext Create a new Next node with the supplied Node.
func NewNext(Node Node) *Next {
	return &Next{
		Node:            Node,
		TransactionChan: make(chan transaction.Transaction),
	}
}
