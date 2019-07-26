package output

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//------------------------------------------------------------------------------

// Output Base used for outputting data to external destination
type Output interface {
	// InputTransactionChan returns a channel used to send transactions for saving.
	SetTransactionChan(<-chan transaction.Transaction)

	component.Base
}
