package pipeline

import (
	"sync"

	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/pipeline/persistence"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

//Pipelines Holds the recursive tree of Nodes and their next nodes, etc
type Pipeline struct {
	name           string
	hash           string
	Root           *node.Next
	receiveJobChan <-chan job.Job
	NodeMap        map[string]*node.Node
	bucket         persistence.Bucket
	activeJobs     sync.WaitGroup
	Logger         zap.SugaredLogger
}

//Start starts the pipeline and start accepting Input
func (p *Pipeline) Start() error {

	// start pipeline nodes
	err := p.Root.Start()
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
func (p *Pipeline) Stop() error {
	// Wait all running jobs to return
	p.activeJobs.Wait()

	//Stop
	err := p.Root.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) handleJob(Job job.Job) {
	responseChan := make(chan response.Response)
	p.activeJobs.Add(1)
	p.Root.JobChan <- job.Job{
		Payload:      Job.Payload,
		Data:         Job.Data,
		ResponseChan: responseChan,
		Context:      Job.Context,
	}
	Job.ResponseChan <- <-responseChan
	p.activeJobs.Done()
}

//recoverAsyncJobs checks pipeline's persisted unfinished jobs and re-apply them
func (p *Pipeline) recoverAsyncJobs() error {

	JobsList, err := p.bucket.GetAllAsyncJobs()
	if err != nil {
		p.Logger.Infow("error occurred while reading in-disk jobs", "error", err.Error())
		return err
	}
	if len(JobsList) <= 0 {
		return nil
	}

	wg := sync.WaitGroup{}

	p.Logger.Infof("re-applying %d async requests found", len(JobsList))
	for _, Job := range JobsList {
		wg.Add(1)
		go func(Job job.Async) {
			defer wg.Done()
			// TODO make this respect pipeline overall concurrent factor
			// Do the Job
			p.reapplyAsyncJob(Job)

			// Delete it from bucket
			err := p.bucket.DeleteAsyncJob(&Job)
			if err != nil {
				p.Logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
			}
		}(Job)
	}

	//cleanup after all jobs are done
	go func() {
		wg.Wait()
		p.bucket.Cleanup()
		p.Logger.Info("finished processing jobs in persistent queue")
	}()

	return nil
}

func (p *Pipeline) reapplyAsyncJob(asyncJob job.Async) {

	p.NodeMap[asyncJob.Node].Process(asyncJob.Job)

	// Wait Response
	response := <-asyncJob.JobResponseChan

	// log progress
	if !response.Ack {
		if response.Error != nil {
			p.Logger.Warnw("an async request that are re-done failed", "error", response.Error)
		} else if response.AckErr != nil {
			p.Logger.Warnw("an async request that are re-done was dropped", "reason", response.AckErr)
		}
	}

}

func (p *Pipeline) createAsyncJob(nodeName string, j job.Job) (*job.Async, error) {

	asyncJob, err := p.bucket.CreateAsyncJob(nodeName, j)
	if err != nil {
		return nil, err
	}

	// --------------------------------------------------------------

	go p.receiveAsyncResponse(asyncJob)

	return asyncJob, nil
}

func (p *Pipeline) receiveAsyncResponse(asyncJob *job.Async) {

	response := <-asyncJob.JobResponseChan
	if response.Error != nil {
		p.Logger.Errorw("error occurred when processing an async request", "error", response.Error.Error())
	}

	// Delete Entry from Repository
	err := p.bucket.DeleteAsyncJob(asyncJob)
	if err != nil {
		p.Logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
	}
}
