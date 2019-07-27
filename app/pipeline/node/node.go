package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

type createAsyncFunc func(nodeName string, j job.Job) (*job.Async, error)

//Node A Node in the pipeline
type Node struct {
	Name           string
	async          bool
	nexts          []Next
	core           core
	createAsyncJob createAsyncFunc
	resource       *component.Resource
	logger         zap.SugaredLogger
	activeJobs     sync.WaitGroup
	receiveJobChan <-chan job.Job
}

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

	go func() {
		for t := range n.receiveJobChan {
			go n.handleJob(t)
		}
	}()

	return nil
}

//Stop Stop this Node and stop all its next nodes.
func (n *Node) Stop() error {
	//wait  jobs to finish
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

// Process process according to its type stream/bytes
func (n *Node) Process(t job.Job) {
	// Start Job according to process payload core
	switch t.Payload.(type) {
	case payload.Bytes:
		n.core.process(t)
	case payload.Stream:
		n.core.processStream(t)
	default:
		// This theoretically shouldn't happen
		t.ResponseChan <- response.Error(fmt.Errorf("invalid process payload type, must be payload.Bytes or payload.Stream"))
	}
}

func (n *Node) handleJob(t job.Job) {

	// if Node is set async, convert to async process
	if n.async {

		async, err := n.createAsyncJob(n.Name, t)
		if err != nil {
			t.ResponseChan <- response.Error(err)
			return
		}

		// Respond to Awaiting sender as now the new process is gonna be handled by Async Manager
		t.ResponseChan <- response.ACK

		t = async.Job
	}

	n.activeJobs.Add(1)
	n.Process(t)
	n.activeJobs.Done()
}

func (n *Node) sendNextsStream(ctx context.Context, writerCloner mirror.Cloner, data payload.Data) chan response.Response {
	responseChan := make(chan response.Response, len(n.nexts))

	for _, next := range n.nexts {
		// Copy new map
		newData := make(payload.Data, len(data))
		for key := range data {
			newData[key] = data[key]
		}

		next.JobChan <- job.Job{
			Payload:      writerCloner.Clone(),
			Data:         newData,
			Context:      ctx,
			ResponseChan: responseChan,
		}
	}
	return responseChan
}

func (n *Node) sendNexts(ctx context.Context, output payload.Bytes, data payload.Data) chan response.Response {
	responseChan := make(chan response.Response, len(n.nexts))

	for _, next := range n.nexts {
		// Copy new map
		newData := make(payload.Data, len(data))
		for key := range data {
			newData[key] = data[key]
		}

		next.JobChan <- job.Job{
			Payload:      output,
			Data:         newData,
			Context:      ctx,
			ResponseChan: responseChan,
		}
	}
	return responseChan
}

func (n *Node) waitResponses(responseChan chan response.Response) response.Response {
	////////////////////////////////////////////
	// receive from next nodes
	count, total := 0, len(n.nexts)
	Response := response.Response{}

	for ; count < total; count++ {
		Response = <-responseChan
		if !Response.Ack {
			break
		}
	}

	return Response
}
