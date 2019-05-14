package wrapper

import (
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/pkg/component"
)

// Input Wraps an Input Plugin Instance
type Input struct {
	component.Input
	node.Resource
}

// Processor wraps a processor Plugin Instance
type Processor struct {
	component.ProcessorBase
	node.Resource
}

// Output Wraps and Input Plugin Instance
type Output struct {
	component.Output
	node.Resource
}
