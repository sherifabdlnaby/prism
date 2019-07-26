package dummy

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sherifabdlnaby/prism/pkg/component"
	cfg "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

// Dummy Input that read a file from root just for testing.
type Dummy struct {
	config   config
	jobs     chan job.Input
	stopChan chan struct{}
	logger   zap.SugaredLogger
	wg       sync.WaitGroup
}

// NewComponent Return a new Base
func NewComponent() component.Base {
	return &Dummy{}
}

// JobChan Return Job Chan used to send job to this Base
func (d *Dummy) JobChan() <-chan job.Input {
	return d.jobs
}

// Init Initializes Plugin
func (d *Dummy) Init(config cfg.Config, logger zap.SugaredLogger) error {
	var err error

	d.config = *defaultConfig()
	err = config.Populate(&d.config)
	if err != nil {
		return err
	}

	d.config.filename, err = config.NewSelector(d.config.FileName)
	if err != nil {
		return err
	}

	d.config.pipeline, err = config.NewSelector(d.config.Pipeline)
	if err != nil {
		return err
	}

	d.jobs = make(chan job.Input)
	d.stopChan = make(chan struct{}, 1)
	d.logger = logger
	return nil
}

// Start Starts Plugin
func (d *Dummy) Start() error {
	d.logger.Debugf("Started Test Input, Sending %d requests!", d.config.Count)

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		flag := false

		for i := 0; i < d.config.Count; i++ {

			select {
			case <-d.stopChan:
				d.logger.Debugw("closing...")
				return
			default:
				d.wg.Add(1)
				go func(i int) {
					defer d.wg.Done()

					// create context
					ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.config.Timeout)*time.Millisecond)
					defer cancel()

					// payloadData (request params)
					payloadData := payload.Data{
						"count": i + 1,
					}

					// get selectors
					pipeline, err := d.config.pipeline.Evaluate(payloadData)
					if err != nil {
						d.logger.Debugw("Error in dummy: ", zap.Error(err))
						return
					}

					filename, err := d.config.filename.Evaluate(payloadData)
					if err != nil {
						d.logger.Debugw("Error in dummy: ", zap.Error(err))
						return
					}

					payloadData["_filename"] = filepath.Base(filename)[0 : len(filepath.Base(filename))-len(filepath.Ext(filepath.Base(filename)))]

					// Lookup Image Data
					reader, err := os.Open(filename)
					if err != nil {
						d.logger.Debugw("Error in dummy: ", zap.Error(err))
						return
					}

					// Send
					responseChan := make(chan response.Response)
					if flag {
						bytes, _ := ioutil.ReadAll(reader)
						// Send Job
						d.logger.Debugw("SENDING A JOB (bytes)....", "ID", i)
						d.jobs <- job.Input{
							Job: job.Job{
								Payload:      payload.Bytes(bytes),
								Data:         payloadData,
								ResponseChan: responseChan,
								Context:      ctx,
							},
							PipelineTag: pipeline,
						}
					} else {
						d.logger.Debugw("SENDING A JOB (stream)...", "ID", i)
						d.jobs <- job.Input{
							Job: job.Job{
								Payload:      reader,
								Data:         payloadData,
								ResponseChan: responseChan,
								Context:      ctx,
							},
							PipelineTag: pipeline,
						}
					}

					// alternate between stream/data
					//flag = !flag

					// Wait Job
					response := <-responseChan

					d.logger.Debugw("RECEIVED RESPONSE.", "ID", i, "ack", response.Ack, "error", response.Error, "AckErr", response.AckErr)
				}(i)

				time.Sleep(time.Millisecond * time.Duration(d.config.Tick))
			}
		}
	}()

	return nil
}

// Stop closes the plugin gracefully
func (d *Dummy) Stop() error {
	d.stopChan <- struct{}{}
	d.wg.Wait()
	close(d.jobs)
	return nil
}
