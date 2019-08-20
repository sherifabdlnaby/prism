package pipeline

import (
	"sync"

	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/pkg/job"
)

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
			p.recoverAsyncJob(&Job)
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

func (p *pipeline) recoverAsyncJob(asyncJob *job.Async) {
	p.startAsyncJob(asyncJob)
	p.handleJob(asyncJob.Job, node.ID(asyncJob.NodeID))
}
