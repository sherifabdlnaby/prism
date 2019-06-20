package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sherifabdlnaby/prism/pkg/component"
	cfg "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

// Webserver take input from HTTP requests
// TODO A lot of things to be added later: URL Signature key, API key..etc
type Webserver struct {
	config       config
	Transactions chan transaction.InputTransaction
	logger       zap.SugaredLogger
	Server       *http.Server
}

//NewComponent returns a new component of type HTTP plugin.
func NewComponent() component.Component {
	return &Webserver{}

}

//InputTransactionChan return transaction channel used to send transactions on.
func (w *Webserver) InputTransactionChan() <-chan transaction.InputTransaction {
	return w.Transactions
}

//Init initializes the web-server and read its config
func (w *Webserver) Init(config cfg.Config, logger zap.SugaredLogger) error {
	var err error

	w.config = *defaultConfig()
	err = config.Populate(&w.config)
	if err != nil {
		return err
	}

	// Init Dynamic Selectors
	for key, value := range w.config.Paths {
		value.pipelineSelector, err = config.NewSelector(value.Pipeline)
		if err != nil {
			return err
		}
		w.config.Paths[key] = value
	}
	w.Transactions = make(chan transaction.InputTransaction)
	w.logger = logger

	w.buildServer()

	return nil
}

// Start : starts the server and serve requests
func (w *Webserver) Start() error {

	// listenAndServe the server
	go func() {
		w.logger.Infof("Http server listening at %d!", w.config.Port)
		err := w.listenAndServe()
		if err != nil && err != http.ErrServerClosed {
			w.logger.Errorw(fmt.Sprintf("webserver listening at port [%v] stopped", w.config.Port), "error", err.Error())
		}
	}()

	return nil
}

//Close : graceful shutdown.
func (w *Webserver) Close() error {
	w.logger.Infof("gracefully shutting down http server at %d...", w.config.Port)

	err := w.Server.Shutdown(context.Background())
	if err != nil {
		return err
	}

	close(w.Transactions)
	return nil
}
