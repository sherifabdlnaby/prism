package registry

import (
	dummyinput "github.com/sherifabdlnaby/prism/internal/input/dummy"
	"github.com/sherifabdlnaby/prism/internal/input/http"
	s3 "github.com/sherifabdlnaby/prism/internal/output/amazon-s3"
	"github.com/sherifabdlnaby/prism/internal/output/disk"
	"github.com/sherifabdlnaby/prism/internal/output/mysql"
	dummyprocessor "github.com/sherifabdlnaby/prism/internal/processor/dummy"
	"github.com/sherifabdlnaby/prism/internal/processor/nude"
	"github.com/sherifabdlnaby/prism/internal/processor/validator"
	"github.com/sherifabdlnaby/prism/internal/processor/vips"
	"github.com/sherifabdlnaby/prism/pkg/component"
)

// registered used to key map components names -> constructors.
var registered = map[string]func() component.Base{
	"dummy_processor": dummyprocessor.NewComponent,
	"dummy_input":     dummyinput.NewComponent,
	"http":            http.NewComponent,
	"disk":            disk.NewComponent,
	"s3":              s3.NewComponent,
	"mysql":           mysql.NewComponent,
	"vips":            vips.NewComponent,
	"validator":       validator.NewComponent,
	"nude_detector":   nude.NewDetector,
	"nude_censor":     nude.NewCensor,
}
