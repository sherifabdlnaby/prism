package mysql

import (
	"database/sql"
	"sync"

	_ "github.com/go-sql-driver/mysql" ///go-sql-driver for mysql
	"github.com/sherifabdlnaby/prism/pkg/component"
	cfg "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

//Mysql struct
type Mysql struct {
	config       config
	Transactions <-chan transaction.Transaction
	stopChan     chan struct{}
	logger       zap.SugaredLogger
	wg           sync.WaitGroup
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &Mysql{}
}

//SetTransactionChan set Transaction chan that this plugin will use to receive transactions
func (m *Mysql) SetTransactionChan(t <-chan transaction.Transaction) {
	m.Transactions = t
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

//WriteOnMysql func takes the transaction
//that to be written on Mysql db
func (m *Mysql) writeOnMysql(db *sql.DB, txn transaction.Transaction) {
	defer m.wg.Done()

	query, err := m.config.query.Evaluate(txn.Data)
	if err != nil {
		txn.ResponseChan <- response.Error(err)
		return
	}
	stmtQuery, err := db.Prepare(query)
	if err != nil {
		txn.ResponseChan <- response.Error(err)
		return
	}
	defer stmtQuery.Close()

	_, err = stmtQuery.Exec()
	if err != nil {
		txn.ResponseChan <- response.Error(err)
		return
	}

	txn.ResponseChan <- response.Ack()
}

// Start the plugin and be ready for taking transactions
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
		for txn := range m.Transactions {
			m.wg.Add(1)
			go m.writeOnMysql(db, txn)
		}
	}()
	return nil
}

//Close func Send a close signal to stop chan
// to stop taking transactions and Close everything safely
func (m *Mysql) Close() error {
	m.stopChan <- struct{}{}
	m.wg.Wait()
	close(m.stopChan)
	return nil
}
