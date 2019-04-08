package app

import (
	dummy_input "github.com/sherifabdlnaby/prism/internal/input/dummy"
	"github.com/sherifabdlnaby/prism/internal/output/disk"
	dummy_processor "github.com/sherifabdlnaby/prism/internal/processor/dummy"
	"github.com/sherifabdlnaby/prism/pkg/types"
)

type registered struct {
	Name        string
	Constructor func() types.Component
}

//Todo make private when migrate from cmd/prism
var Registry = map[string]registered{
	"dummy_processor": {"dummy_processor", dummy_processor.NewComponent},
	"dummy_input":     {"dummy_input", dummy_input.NewComponent},
	"disk":            {"disk", disk.NewComponent},
}
