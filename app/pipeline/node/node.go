package node

import (
	"fmt"
	"sync"

	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

type createAsyncFunc func(nodeID ID, j job.Job) (*job.Job, error)

//Node A Node in the pipeline
type Node struct {
	ID             ID
	async          bool
	nexts          []Next
	core           core
	createAsyncJob createAsyncFunc
	resource       *component.Resource
	logger         zap.SugaredLogger
	activeJobs     sync.WaitGroup
	receiveJobChan <-chan job.Job
}

type ID string

// Start starts this Node and all its next nodes to start receiving jobs
// By starting all next nodes, start async request handler, and start receiving jobs
func (n *Node) Start() error {
	// Start next nodes
	for _, value := range n.nexts {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	go n.serve()

	return nil
}

//Stop Stop this Node and stop all its next nodes.
func (n *Node) Stop() error {
	//wait jobs to finish
	n.activeJobs.Wait()

	for _, value := range n.nexts {
		// close this next-node chan
		close(value.JobChan)

		// tell this next-node to stop which in turn will close all its next(s) too.
		err := value.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) serve() {
	for j := range n.receiveJobChan {
		if n.async {
			go n.HandleJobAsync(j)
		} else {
			go n.HandleJob(j)
		}
	}
}

func (n *Node) HandleJobAsync(j job.Job) {

	// if Node is set async, convert to async process
	if n.async {

		asyncJob, err := n.createAsyncJob(n.ID, j)
		if err != nil {
			j.ResponseChan <- response.Error(err)
			return
		}

		j = *asyncJob
	}

	n.HandleJob(j)
}

func (n *Node) HandleJob(j job.Job) {
	n.activeJobs.Add(1)
	n.process(j)
	n.activeJobs.Done()
}

// process process according to its type stream/bytes
func (n *Node) process(j job.Job) {
	// Start Job according to process payload core
	switch j.Payload.(type) {
	case payload.Bytes:
		n.core.process(j)
	case payload.Stream:
		n.core.processStream(j)
	default:
		// This theoretically shouldn't happen
		j.ResponseChan <- response.Error(fmt.Errorf("invalid process payload type, must be payload.Bytes or payload.Stream"))
	}
}
