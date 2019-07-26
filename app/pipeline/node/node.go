package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/app/pipeline/persistence"

	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

//Node A Node in the pipeline
type Node struct {
	Name           string
	async          bool
	nexts          []Next
	nodeType       Type
	wg             sync.WaitGroup
	resource       *component.Resource
	Logger         zap.SugaredLogger
	persistence    persistence.Persistence
	receiveJobChan chan job.Job //TODO make a receive only for more sanity
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
			n.handleJob(t)
		}
	}()

	return nil
}

//Stop Stop this Node and stop all its next nodes.
func (n *Node) Stop() error {
	//wait async jobs to finish
	n.wg.Wait()

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

//SetJobChan Set the job chan Node will use to receive input
func (n *Node) SetJobChan(tc chan job.Job) {
	n.receiveJobChan = tc
}

//ReceiveJobChan Getter for receiveChan
func (n *Node) ReceiveJobChan() <-chan job.Job {
	return n.receiveJobChan
}

//SetNexts Set this Node's next nodes.
func (n *Node) SetNexts(nexts []Next) {
	n.nexts = nexts
}

//SetAsync Set if this Node is sync/async
func (n *Node) SetAsync(async bool) {
	n.async = async
}

//SetAsync Set if this Node is sync/async
func (n *Node) SetPersistence(p persistence.Persistence) {
	n.persistence = p
}

// ProcessJob job according to its type stream/bytes
func (n *Node) ProcessJob(t job.Job) {
	// Start Job according to job payload Type
	switch t.Payload.(type) {
	case payload.Bytes:
		go n.nodeType.job(t)
	case payload.Stream:
		go n.nodeType.jobStream(t)
	default:
		// This theoretically shouldn't happen
		t.ResponseChan <- response.Error(fmt.Errorf("invalid job payload type, must be payload.Bytes or payload.Stream"))
	}
}

func (n *Node) handleJob(t job.Job) {

	// if Node is set async, convert to async job
	if n.async {

		newJob, err := n.startAsyncJob(t)
		if err != nil {
			t.ResponseChan <- response.Error(err)
			return
		}

		// Respond to Awaiting sender as now the new job is gonna be handled by Async Manager
		t.ResponseChan <- response.ACK

		t = newJob
	}

	n.ProcessJob(t)
}

func (n *Node) startAsyncJob(t job.Job) (job.Job, error) {

	asyncJob, newPayload, err := n.persistence.SaveJob(n.Name, t)
	if err != nil {
		return job.Job{}, nil
	}

	// --------------------------------------------------------------

	newResponseChan := make(chan response.Response)
	n.wg.Add(1)
	go n.receiveAsyncResponse(asyncJob, newResponseChan)

	// since it will be async, sync job context is irrelevant.
	// (we don't want sync nodes -that cancel job context when finishing to avoid ctx leak- to
	// cancel async nodes too )
	t.Context = context.Background()

	// New Payload
	t.Payload = newPayload

	// now actual response is given to asyncResponds that should handle async responds
	t.ResponseChan = newResponseChan

	return t, nil
}

func (n *Node) receiveAsyncResponse(asyncJob *job.Async, newResponseChan chan response.Response) {
	defer n.wg.Done()

	response := <-newResponseChan
	if response.Error != nil {
		n.Logger.Errorw("error occurred when processing an async request", "error", response.Error.Error())
	}

	// ------------------ Clean UP ------------------ //
	err := asyncJob.Finalize()
	if err != nil {
		n.Logger.Errorw("an error occurred while applying finalizing async requests", "error", err.Error())
	}

	// Delete Entry from DB
	err = n.persistence.DeleteJob(asyncJob)
	if err != nil {
		n.Logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
	}
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
