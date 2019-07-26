package input

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

// Input is a type that sends messages as transactions and waits for a
// response back.
type Input interface {
	// InputTransactionChan returns a channel used for consuming transactions from
	// this type.
	InputTransactionChan() <-chan transaction.InputTransaction

	component.Base
}
