package mysql

import (
	"database/sql"
	"sync"

	_ "github.com/go-sql-driver/mysql" ///go-sql-driver for mysql
	"github.com/sherifabdlnaby/prism/pkg/component"
	cfg "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

//Mysql struct
type Mysql struct {
	config   config
	jobChan  <-chan job.Job
	stopChan chan struct{}
	logger   zap.SugaredLogger
	wg       sync.WaitGroup
}

// NewComponent Return a new Base
func NewComponent() component.Base {
	return &Mysql{}
}

//SetJobChan set Job chan that this plugin will use to receive jobs
func (m *Mysql) SetJobChan(t <-chan job.Job) {
	m.jobChan = t
}

//Init func Initialize Mysql output plugin
func (m *Mysql) Init(config cfg.Config, logger zap.SugaredLogger) error {

	err := config.Populate(&m.config)
	if err != nil {
		return err
	}

	m.config.query, err = config.NewSelector(m.config.Query)
	if err != nil {
		return err
	}

	m.stopChan = make(chan struct{})
	m.logger = logger

	return nil
}

//WriteOnMysql func takes the job
//that to be written on Mysql db
func (m *Mysql) writeOnMysql(db *sql.DB, job job.Job) {
	defer m.wg.Done()

	query, err := m.config.query.Evaluate(job.Data)
	if err != nil {
		job.ResponseChan <- response.Error(err)
		return
	}
	stmtQuery, err := db.Prepare(query)
	if err != nil {
		job.ResponseChan <- response.Error(err)
		return
	}
	defer stmtQuery.Close()

	_, err = stmtQuery.Exec()
	if err != nil {
		job.ResponseChan <- response.Error(err)
		return
	}

	job.ResponseChan <- response.Ack()
}

// Start the plugin and be ready for taking jobs
func (m *Mysql) Start() error {
	dataSourceName := m.config.Username + ":" + m.config.Password + "@/" + m.config.DBName
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for Job := range m.jobChan {
			m.wg.Add(1)
			go m.writeOnMysql(db, Job)
		}
	}()
	return nil
}

//Stop func Send a close signal to stop chan
// to stop taking jobs and Stop everything safely
func (m *Mysql) Stop() error {
	m.stopChan <- struct{}{}
	m.wg.Wait()
	close(m.stopChan)
	return nil
}
