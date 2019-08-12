package pipeline

import (
	"sync"

	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/pipeline/persistence"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

//Pipelines Holds the recursive tree of Nodes and their next nodes, etc
type pipeline struct {
	name           string
	hash           string
	root           *node.Next
	resource       component.Resource
	receiveJobChan <-chan job.Job
	nodeMap        map[string]*node.Node
	bucket         persistence.Bucket
	activeJobs     sync.WaitGroup
	logger         zap.SugaredLogger
}

//Start starts the pipeline and start accepting Input
func (p *pipeline) Start() error {

	// start pipeline nodes
	err := p.root.Start()
	if err != nil {
		return err
	}

	go func() {
		for value := range p.receiveJobChan {
			go p.handleJob(value)
		}
	}()

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

func (p *pipeline) ActiveJobs() int {
	return p.resource.Current()
}

func (p *pipeline) handleJob(Job job.Job) {

	p.activeJobs.Add(1)
	err := p.resource.Acquire(Job.Context)
	if err != nil {
		Job.ResponseChan <- response.NoAck(err)
		return
	}

	// -----------------------------------------

	responseChan := make(chan response.Response)
	p.root.JobChan <- job.Job{
		Payload:      Job.Payload,
		Data:         Job.Data,
		ResponseChan: responseChan,
		Context:      Job.Context,
	}
	Job.ResponseChan <- <-responseChan

	// -----------------------------------------

	p.resource.Release()
	p.activeJobs.Done()
}

func (p *pipeline) createAsyncJob(nodeName string, j job.Job) (*job.Async, error) {

	asyncJob, err := p.bucket.CreateAsyncJob(nodeName, j)
	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------------

	p.activeJobs.Add(1)
	p.resource.Acquire(asyncJob.Job.Context)

	go func() {
		p.receiveAsyncResponse(asyncJob)
		p.resource.Release()
		p.activeJobs.Done()
	}()

	return asyncJob, nil
}

// TODO refactor
func (p *pipeline) receiveAsyncResponse(asyncJob *job.Async) {

	response := <-asyncJob.JobResponseChan
	if response.Error != nil {
		p.logger.Errorw("error occurred when processing an async request", "error", response.Error.Error())
	}

	// Delete Entry from Repository
	err := p.bucket.DeleteAsyncJob(asyncJob)
	if err != nil {
		p.logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
	}

}

//recoverAsyncJobs checks pipeline's persisted unfinished jobs and re-apply them
func (p *pipeline) recoverAsyncJobs() error {

	JobsList, err := p.bucket.GetAllAsyncJobs()
	if err != nil {
		p.logger.Infow("error occurred while reading in-disk jobs", "error", err.Error())
		return err
	}
	if len(JobsList) <= 0 {
		return nil
	}

	wg := sync.WaitGroup{}

	p.logger.Infof("re-applying %d async requests found", len(JobsList))
	for _, Job := range JobsList {
		wg.Add(1)
		go func(Job job.Async) {
			defer wg.Done()
			// Do the Job
			p.handleJobAsync(Job)
		}(Job)
	}

	//cleanup after all jobs are done
	go func() {
		wg.Wait()
		p.bucket.Cleanup()
		p.logger.Info("finished processing jobs in persistent queue")
	}()

	return nil
}

func (p *pipeline) handleJobAsync(asyncJob job.Async) {
	p.activeJobs.Add(1)

	err := p.resource.Acquire(asyncJob.Job.Context)
	if err != nil {
		p.logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
	}

	p.nodeMap[asyncJob.Node].Process(asyncJob.Job)

	// Wait Response
	response := <-asyncJob.JobResponseChan

	// log progress
	if !response.Ack {
		if response.Error != nil {
			p.logger.Warnw("an async request that are re-done failed", "error", response.Error)
		} else if response.AckErr != nil {
			p.logger.Warnw("an async request that are re-done was dropped", "reason", response.AckErr)
		}
	}

	// Delete it from bucket
	err = p.bucket.DeleteAsyncJob(&asyncJob)
	if err != nil {
		p.logger.Errorw("an error occurred while deleting temp file ", "error", err.Error())
	}

	p.resource.Release()
	p.activeJobs.Done()
}
