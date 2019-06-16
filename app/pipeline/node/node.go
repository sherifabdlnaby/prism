package node

import (
	"reflect"

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

	//SetNexts Set this Node's next nodes.
	GetInternalType() interface{}
}

//TODO rename
type component interface {
	job(t transaction.Transaction)
	jobStream(t transaction.Transaction)
	getInternalType() interface{}
}

//Next Wraps the next Node plus the channel used to communicate with this Node to send input transactions.
type Next struct {
	Node
	TransactionChan chan transaction.Transaction
	Same            bool
}

//NewNext Create a new Next Node with the supplied Node.
func NewNext(next, parent Node) *Next {

	var same bool

	if parent != nil {
		typ1 := parent.GetInternalType()
		typ2 := next.GetInternalType()
		if typ1 != nil && typ2 != nil {
			name1 := reflect.TypeOf(typ1).Elem().String()
			name2 := reflect.TypeOf(typ2).Elem().String()
			same = name1 == name2
		}
	}

	return &Next{
		Node:            next,
		TransactionChan: make(chan transaction.Transaction),
		Same:            same,
	}
}
