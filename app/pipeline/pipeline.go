package pipeline

import (
	"sync"
	"sync/atomic"

	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/pipeline/persistence"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

//Pipelines Holds the recursive tree of Nodes and their next nodes, etc
type pipeline struct {
	name            string
	hash            string
	root            *node.Next
	resource        component.Resource
	receiveJobChan  <-chan job.Job
	handleAsyncJobs chan *job.Async
	nodeMap         map[node.ID]*node.Node
	bucket          persistence.Bucket
	activeJobs      sync.WaitGroup
	jobsCounter     int32
	logger          zap.SugaredLogger
}

const root node.ID = ""

//Start starts the pipeline and start accepting Input
func (p *pipeline) Start() error {

	// start pipeline nodes
	err := p.root.Start()
	if err != nil {
		return err
	}

	go p.asyncResponsesManager()
	go p.serve()

	return nil
}

// Stop stops the pipeline, that means that any job received on this pipeline after stopping will return
// error response unless re-started again.
func (p *pipeline) Stop() error {

	// Wait all running jobs to return
	p.activeJobs.Wait()

	// Stop
	err := p.root.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (p *pipeline) serve() {
	for value := range p.receiveJobChan {
		go p.handleJob(value, root)
	}
}

func (p *pipeline) handleJob(Job job.Job, nodeID node.ID) {

	p.activeJobs.Add(1)
	atomic.AddInt32(&p.jobsCounter, 1)
	err := p.resource.Acquire(Job.Context)
	if err != nil {
		Job.ResponseChan <- response.NoAck(err)
		return
	}

	// -----------------------------------------

	responseChan := make(chan response.Response, 1)

	// If node name supplied, send job to said node, else root.
	if nodeID == root {
		p.root.HandleJob(job.Job{
			Payload:      Job.Payload,
			Data:         Job.Data,
			ResponseChan: responseChan,
			Context:      Job.Context,
		})
	} else {
		p.nodeMap[nodeID].HandleJob(job.Job{
			Payload:      Job.Payload,
			Data:         Job.Data,
			ResponseChan: responseChan,
			Context:      Job.Context,
		})
	}

	// await response
	Job.ResponseChan <- <-responseChan

	// -----------------------------------------

	p.resource.Release()
	atomic.AddInt32(&p.jobsCounter, -1)
	p.activeJobs.Done()
}

func (p *pipeline) convertToAsync(ID node.ID, j job.Job) (*job.Job, error) {

	asyncJob, err := p.bucket.CreateAsyncJob(ID, j)
	if err != nil {
		return nil, err
	}

	p.startAsyncJob(asyncJob)

	// Respond to Awaiting sender as now the new process is gonna be handled by Async Manager
	j.ResponseChan <- response.ACK

	return &asyncJob.Job, nil
}

func (p *pipeline) startAsyncJob(asyncJob *job.Async) {
	// Acquire Resources
	p.activeJobs.Add(1)

	atomic.AddInt32(&p.jobsCounter, 1)

	// Send Async JOB to async handler to deal with its response
	p.handleAsyncJobs <- asyncJob
}

func (p *pipeline) asyncResponsesManager() {
	for asyncJob := range p.handleAsyncJobs {
		go p.waitAndFinalizeAsyncJob(*asyncJob)
	}
}

func (p *pipeline) waitAndFinalizeAsyncJob(asyncJob job.Async) {
	defer func() {
		atomic.AddInt32(&p.jobsCounter, -1)
		p.activeJobs.Done()
	}()

	response := <-asyncJob.JobResponseChan
	if response.Error != nil {
		p.logger.Errorw("error occurred when processing an async request", "error", response.Error.Error())
	}

	// Delete Entry from Repository
	err := p.bucket.DeleteAsyncJob(&asyncJob)
	if err != nil {
		p.logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
	}

	p.logger.Debug("DONE WITH ", asyncJob.Data["count"])
}

func (p *pipeline) ActiveJobs() int {
	return int(p.jobsCounter)
}
