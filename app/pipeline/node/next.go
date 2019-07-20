package node

import "github.com/sherifabdlnaby/prism/pkg/transaction"

//Next Wraps the next Node plus the channel used to communicate with this Node to send input transactions.
type Next struct {
	*Node
	TransactionChan chan transaction.Transaction
}

//NewNext Create a new Next Node with the supplied Node.
func NewNext(node *Node) *Next {
	transactionChan := make(chan transaction.Transaction)

	// gives the next's Node its InputTransactionChan, now owner of the 'next' owns closing the chan.
	node.SetTransactionChan(transactionChan)

	return &Next{
		Node:            node,
		TransactionChan: transactionChan,
	}
}
