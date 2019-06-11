package registery

import (
	dummyinput "github.com/sherifabdlnaby/prism/internal/input/dummy"
	"github.com/sherifabdlnaby/prism/internal/output/disk"
	dummyprocessor "github.com/sherifabdlnaby/prism/internal/processor/dummy"
	"github.com/sherifabdlnaby/prism/internal/processor/vips"
	"github.com/sherifabdlnaby/prism/pkg/component"
)

// registered used to key map components names -> constructors.
var registered = map[string]func() component.Component{
	"dummy_processor": dummyprocessor.NewComponent,
	"dummy_input":     dummyinput.NewComponent,
	"disk":            disk.NewComponent,
	//"s3":              s3.NewComponent,
	"vips": vips.NewComponent,
}
