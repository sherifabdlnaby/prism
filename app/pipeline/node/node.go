package node

import "github.com/sherifabdlnaby/prism/pkg/transaction"

// TODO refactor nodes to make code more reusable

//Node A node wraps components and manage receiving transactions and forwarding transactions to next nodes.
type Node interface {
	Start() error
	Stop() error
	SetTransactionChan(tc <-chan transaction.Transaction)
}

//Root Wraps the next node plus its attributes.
type Next struct {
	Node
	TransactionChan chan<- transaction.Transaction
}
