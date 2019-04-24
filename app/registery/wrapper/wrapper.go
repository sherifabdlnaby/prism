package wrapper

import (
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component"
)

// Input Wraps an Input Plugin Instance
type Input struct {
	component.Input
	resource.Resource
}

// Processor wraps a processor Plugin Instance
type Processor struct {
	component.ProcessorBase
	resource.Resource
}

// Output Wraps and Input Plugin Instance
type Output struct {
	component.Output
	resource.Resource
}
