package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"sync"

	//"github.com/sherifabdlnaby/prism/pkg/bufferspool"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	//"github.com/sherifabdlnaby/prism/pkg/payload"
	//"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

//Mysql struct
type Mysql struct {
	Username     config.Value
	Password     config.Value
	DBName       config.Value
	Query        config.Value
	TypeCheck    bool
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
func (m *Mysql) Init(config config.Config, logger zap.SugaredLogger) error {
	var err error

	m.Username, err = config.Get("username", nil)
	if err != nil {
		return err
	}
	m.Password, err = config.Get("password", nil)
	if err != nil {
		return err
	}
	m.DBName, err = config.Get("dbname", nil)
	if err != nil {
		return err
	}
	m.Query, err = config.Get("query", nil)
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
	dataSourceName := m.Username.Get().String() + ":" + m.Password.Get().String() + "@/" + m.DBName.Get().String()
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
	//d.wg.Wait()
	return nil
}
