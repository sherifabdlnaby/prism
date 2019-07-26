package pipeline

import (
	"context"
	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/pipeline/persistence"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
	"io"
	"sync"
)

//Pipelines Holds the recursive tree of Nodes and their next nodes, etc
type Pipeline struct {
	name           string
	hash           string
	registry       component.Registry
	Root           *node.Next
	NodeMap        map[string]*node.Node
	receiveJobChan <-chan job.Job
	persistence    persistence.Persistence
	wg             sync.WaitGroup
	config         config.Pipeline
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
			go p.process(value)
		}
	}()

	return nil
}

// Stop stops the pipeline, that means that any job received on this pipeline after stopping will return
// error response unless re-started again.
func (p *Pipeline) Stop() error {
	// Wait all running jobs to return
	p.wg.Wait()

	//Stop
	err := p.Root.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) process(Job job.Job) {
	p.wg.Add(1)
	responseChan := make(chan response.Response)
	p.Root.JobChan <- job.Job{
		Payload:      Job.Payload,
		Data:         Job.Data,
		ResponseChan: responseChan,
		Context:      Job.Context,
	}
	Job.ResponseChan <- <-responseChan
	p.wg.Done()
}

//recoverAsync checks pipeline's persisted unfinished jobs and re-apply them
func (p *Pipeline) recoverAsync() error {

	JobsList, err := p.persistence.GetAllJobs()
	if err != nil {
		p.Logger.Infow("error occurred while reading in-disk jobs", "error", err.Error())
		return err
	}

	p.Logger.Infof("re-applying %d async requests found", len(JobsList))
	for _, asyncJobs := range JobsList {
		p.applyAsyncJob(asyncJobs)
	}

	return nil
}

func (p *Pipeline) applyAsyncJob(asyncJob job.Async) {

	// Send Job to the Async Node
	responseChan := make(chan response.Response)
	p.NodeMap[asyncJob.Node].ProcessJob(job.Job{
		Payload:      io.Reader(asyncJob.TmpFile),
		Data:         asyncJob.Data,
		Context:      context.Background(),
		ResponseChan: responseChan,
	})

	// Wait Response
	response := <-responseChan

	// log progress
	if !response.Ack {
		if response.Error != nil {
			p.Logger.Warnw("an async request that are re-done failed", "error", response.Error)
		} else if response.AckErr != nil {
			p.Logger.Warnw("an async request that are re-done was dropped", "reason", response.AckErr)
		}
	}

	// ------------------ Clean UP ------------------ //
	err := asyncJob.Finalize()
	if err != nil {
		p.Logger.Errorw("an error occurred while applying finalizing async requests", "error", err.Error())
	}

	// Delete Entry from DB
	err = p.persistence.DeleteJob(&asyncJob)
	if err != nil {
		p.Logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
	}
}
