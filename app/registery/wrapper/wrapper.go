package wrapper

import (
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"go.uber.org/zap"
)

// Input Wraps an Input Plugin Instance
type Input struct {
	component.Input
	resource.Resource
	Logger zap.SugaredLogger
}

// Processor wraps a processor Plugin Instance
type Processor struct {
	component.ProcessorBase
	resource.Resource
	Logger zap.SugaredLogger
}

// Output Wraps and Input Plugin Instance
type Output struct {
	component.Output
	resource.Resource
	Logger zap.SugaredLogger
}
