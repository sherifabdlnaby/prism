package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

// The Webserver receiving input.
//A lot of things to be added later: URL Signature key, HTTPs (key,cert), API key..etc
type Webserver struct {
	config       Config
	Transactions chan transaction.InputTransaction
	logger       zap.SugaredLogger
	Server       *http.Server
}

type Config struct {
	Port     int
	CertFile string
	KeyFile  string
	Paths    map[string]path
}
type path struct {
	Pipeline         string
	pipelineSelector config.Selector
}

func DefaultConfig() *Config {
	return &Config{}
}

//Newcomponent returns a new component.
func NewComponent() component.Component {
	return &Webserver{}

}

//TransactionChan return transaction channel.
func (w *Webserver) InputTransactionChan() <-chan transaction.InputTransaction {
	return w.Transactions
}

//Init initializes webserver
func (w *Webserver) Init(config config.Config, logger zap.SugaredLogger) error {

	var err error
	w.config = *DefaultConfig()
	err = config.Populate(&w.config)
	if err != nil {
		return err
	}

	for k := range w.config.Paths {
		//Go is a weird language. {cannot assign to struct field in map}
		var tmp = w.config.Paths[k]
		tmp.pipelineSelector, err = config.NewSelector(w.config.Paths[k].Pipeline)
		if err != nil {
			return err
		}
		w.config.Paths[k] = tmp
	}

	w.Transactions = make(chan transaction.InputTransaction)
	w.logger = logger
	return nil
}

// Start : starts the server and waits for requests.
func (w *Webserver) Start() error {
	w.logger.Info(fmt.Sprintf("Webserver listening at %d!", w.config.Port))

	// Start the server
	go Server(w)

	return nil
}

//Close : graceful shutdown.
func (w *Webserver) Close() error {
	w.logger.Infof("gracefully shutting down http server at port: %i...", w.config.Port)

	err := w.Server.Shutdown(context.Background())
	if err != nil {
		return err
	}

	close(w.Transactions)
	return nil
}
