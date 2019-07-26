package input

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/job"
)

// Input is a type that sends messages as jobs and waits for a
// response back.
type Input interface {
	// JobChan returns a channel used for consuming jobs from
	// this type.
	JobChan() <-chan job.Input

	component.Base
}
