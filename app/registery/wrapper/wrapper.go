package wrapper

import (
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

// Input Wraps an Input Plugin Instance
type Input struct {
	component.Input
	Resource resource.Resource
}

// Processor wraps a processor Plugin Instance
type Processor struct {
	component.ProcessorBase
	Resource resource.Resource
}

// Output Wraps and Input Plugin Instance
type Output struct {
	component.Output
	Resource        resource.Resource
	TransactionChan chan transaction.Transaction
}
