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
	persistence    persistence.Persistence
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

	JobsList, err := p.persistence.GetAllAsyncJobs()
	if err != nil {
		p.Logger.Infow("error occurred while reading in-disk jobs", "error", err.Error())
		return err
	}

	p.Logger.Infof("re-applying %d async requests found", len(JobsList))
	for _, asyncJobs := range JobsList {
		go func() {
			// TODO make this respect pipeline overall concurrent factor
			// TODO make this concurrent
			p.reapplyAsyncJob(asyncJobs)
		}()
	}

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

	// ------------------ Clean UP ------------------ //

	// Delete Entry from DB
	err := p.persistence.DeleteAsyncJob(&asyncJob)
	if err != nil {
		p.Logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
	}
}

func (p *Pipeline) createAsyncJob(nodeName string, j job.Job) (*job.Async, error) {

	asyncJob, err := p.persistence.CreateAsyncJob(nodeName, j)
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

	// Delete Entry from DB
	err := p.persistence.DeleteAsyncJob(asyncJob)
	if err != nil {
		p.Logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
	}
}
