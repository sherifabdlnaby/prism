package output

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/job"
)

//------------------------------------------------------------------------------

// Output Base used for outputting data to external destination
type Output interface {
	// JobChan returns a channel used to send jobs for saving.
	SetJobChan(<-chan job.Job)

	component.Base
}
