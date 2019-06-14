package mysql

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
	"sync"
	"time"
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

type DatabaseConfig struct {
	DBName               string `toml:"dbname"`
	Host                 string `toml:"host"`
	Port                 int    `toml:"port"`
	User                 string `toml:"user"`
	Password             string `toml:"password"`
	Sslmode              string `toml:"sslmode"`
	ShowLog              bool
	DataSaveDir          string
	DataFileSaveLoopSize int
	MaxIdleConns         int `toml:"max_idle_conns"`
	MaxOpenConns         int `toml:"max_open_conns"`
	MaxLifeTime          int `toml:"max_life_time"`
	Timeout              int `toml:"timeout"`
	RTimeout             int `toml:"rtimeout"`
	WTimeout             int `toml:"wtimeout"`
}

func (c DatabaseConfig) MySQLSource() string {
	params := make(map[string]string, 0)
	params["charset"] = "utf8mb4"
	cfg := mysql.Config{}
	cfg.User = c.User
	cfg.Passwd = c.Password
	cfg.DBName = c.DBName
	cfg.ParseTime = true
	cfg.Collation = "utf8mb4_unicode_ci"
	cfg.Params = params
	cfg.Loc, _ = time.LoadLocation("Asia/Chongqing")
	cfg.Timeout = time.Duration(c.Timeout) * time.Second
	cfg.MultiStatements = true
	cfg.ReadTimeout = time.Duration(c.RTimeout) * time.Second
	cfg.WriteTimeout = time.Duration(c.WTimeout) * time.Second
	return cfg.FormatDSN()
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
