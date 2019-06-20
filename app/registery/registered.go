package registery

import (
	dummyinput "github.com/sherifabdlnaby/prism/internal/input/dummy"
	s3 "github.com/sherifabdlnaby/prism/internal/output/amazon-s3"
	"github.com/sherifabdlnaby/prism/internal/output/disk"
	"github.com/sherifabdlnaby/prism/internal/output/mysql"
	"github.com/sherifabdlnaby/prism/internal/processor/detectors"
	dummyprocessor "github.com/sherifabdlnaby/prism/internal/processor/dummy"
	"github.com/sherifabdlnaby/prism/internal/processor/nude"
	"github.com/sherifabdlnaby/prism/internal/processor/vips"
	"github.com/sherifabdlnaby/prism/pkg/component"
)

// registered used to key map components names -> constructors.
var registered = map[string]func() component.Component{
	"dummy_processor": dummyprocessor.NewComponent,
	"dummy_input":     dummyinput.NewComponent,
	"disk":            disk.NewComponent,
	"s3":              s3.NewComponent,
	"mysql":           mysql.NewComponent,
	"vips":            vips.NewComponent,
	"face_detector":   detectors.NewFaceDetecor,
	"nude_detector":   nude.NewDetector,
	"nude_censor":     nude.NewCensor,
}
