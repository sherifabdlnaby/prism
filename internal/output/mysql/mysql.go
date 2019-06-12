package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
	"sync"
)

//Mysql struct
type Mysql struct {
	config       Config
	Transactions <-chan transaction.Transaction
	stopChan     chan struct{}
	logger       zap.SugaredLogger
	wg           sync.WaitGroup
}

//Config struct
type Config struct {
	Username string `mapstructure:"username" validate:"required"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name" validate:"required"`
	Query    string `mapstructure:"query" validate:"required"`
	query    config.Selector
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
func (m *Mysql) Init(config config.Config, logger zap.SugaredLogger) error {

	var err error

	err = config.Populate(&m.config)
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
func (m *Mysql) writeOnMysql(txn transaction.Transaction) {
	defer m.wg.Done()

}

// Start the plugin and be ready for taking transactions
func (m *Mysql) Start() error {
	dataSourceName := m.config.Username + ":" + m.config.Password + "@/" + m.config.DBName
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		return err
	}
	fmt.Println("aloo")
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for txn := range m.Transactions {
			m.wg.Add(1)
			go m.writeOnMysql(txn)
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
