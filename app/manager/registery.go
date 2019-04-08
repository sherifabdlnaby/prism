package manager

import (
	dummyinput "github.com/sherifabdlnaby/prism/internal/input/dummy"
	"github.com/sherifabdlnaby/prism/internal/output/disk"
	dummyprocessor "github.com/sherifabdlnaby/prism/internal/processor/dummy"
	"github.com/sherifabdlnaby/prism/pkg/types"
)

// registered used to key map components names -> constructors.
var registered = map[string]func() types.Component{
	"dummy_processor": dummyprocessor.NewComponent,
	"dummy_input":     dummyinput.NewComponent,
	"disk":            disk.NewComponent,
}
