package disk

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/sherifabdlnaby/prism/pkg/component"
	cfg "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

//Disk struct
type Disk struct {
	config   config
	jobsChan <-chan job.Job
	stopChan chan struct{}
	logger   zap.SugaredLogger
	wg       sync.WaitGroup
}

// NewComponent Return a new Base
func NewComponent() component.Base {
	return &Disk{}
}

//SetJobChan set Job chan that this plugin will use to receive jobs
func (d *Disk) SetJobChan(t <-chan job.Job) {
	d.jobsChan = t
}

//Init func Initialize the disk output plugin
func (d *Disk) Init(config cfg.Config, logger zap.SugaredLogger) error {
	var err error

	d.config = *defaultConfig()
	err = config.Populate(&d.config)
	if err != nil {
		return err
	}

	d.config.filepath, err = config.NewSelector(d.config.FilePath)
	if err != nil {
		return err
	}

	d.stopChan = make(chan struct{})
	d.logger = logger

	return nil
}

// Start the plugin and be ready for taking jobs
func (d *Disk) Start() error {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for Job := range d.jobsChan {
			d.wg.Add(1)
			go d.writeOnDisk(Job)
		}
	}()
	return nil
}

//Stop func Send a close signal to stop chan
// to stop taking jobs and Stop everything safely
func (d *Disk) Stop() error {
	d.wg.Wait()
	return nil
}

//WriteOnDisk func takes the job
//that to be written on the disk
func (d *Disk) writeOnDisk(Job job.Job) {
	defer d.wg.Done()

	filePath, err := d.config.filepath.Evaluate(Job.Data)
	if err != nil {
		Job.ResponseChan <- response.Error(err)
		return
	}

	dir := filepath.Dir(filePath)

	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
	}

	if err != nil {
		Job.ResponseChan <- response.Error(err)
		return
	}

	switch Payload := Job.Payload.(type) {
	case payload.Bytes:
		err = ioutil.WriteFile(filePath, Payload, d.config.Permission)
	case payload.Stream:
		err = writeFileFromStream(filePath, Payload, d.config.Permission)
	}

	if err != nil {
		// send response
		Job.ResponseChan <- response.Error(err)
		return
	}

	// send response
	Job.ResponseChan <- response.Ack()
}

func writeFileFromStream(filename string, reader io.Reader, perm os.FileMode) error {

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	defer func() {
		err := f.Close()
		if err != nil {
			log.Print("error while closing file! error:", err)
		}
	}()

	if err != nil {
		return err
	}

	_, err = io.Copy(f, reader)
	if err != nil {
		return err
	}

	return nil
}
