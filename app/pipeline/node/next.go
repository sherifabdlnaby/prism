package node

import "github.com/sherifabdlnaby/prism/pkg/job"

//Next Wraps the next Node plus the channel used to communicate with this Node to send input jobs.
type Next struct {
	*Node
	JobChan chan job.Job
}
