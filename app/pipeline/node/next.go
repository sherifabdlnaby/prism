package node

import "github.com/sherifabdlnaby/prism/pkg/job"

//Next Wraps the next Node plus the channel used to communicate with this Node to send input jobs.
type Next struct {
	*Node
	JobChan chan job.Job
}

//NewNext Create a new Next Node with the supplied Node.
func NewNext(node *Node) *Next {
	JobChan := make(chan job.Job)

	// gives the next's Node its JobChan, now owner of the 'next' owns closing the chan.
	node.SetJobChan(JobChan)

	return &Next{
		Node:    node,
		JobChan: JobChan,
	}
}
