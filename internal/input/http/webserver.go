package http

import (
	"context"
	"fmt"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	response2 "github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
	"io"
	"net/http"
	"sync"
)

// The Webserver receiving input.
//A lot of things to be added later: URL Signature key, HTTPs (key,cert), API key..etc
type Webserver struct {
	Transactions chan transaction.InputTransaction
	stopChan     chan struct{}
	logger       zap.SugaredLogger
	wg           sync.WaitGroup
	port         int
	CertFile     string
	KeyFile      string
	Server       *http.Server
	Requests     chan http.Request
}

//Newcomponent returns a new component.
func NewComponent() component.Component {
	return &Webserver{}
}

//TransactionChan return transaction channel.
func (w *Webserver) TransactionChan() <-chan transaction.InputTransaction {
	return w.Transactions
}

//Init initializes webserver
func (w *Webserver) Init(config config.Config, logger zap.SugaredLogger) error {
	p, err := config.Get("port", nil)
	if err != nil {
		return err
	}
	w.port = p.Get().Int()
	//w.KeyFile = config["KeyFile"].(string)
	//w.CertFile = config["CertFile"].(string)

	w.Transactions = make(chan transaction.InputTransaction, 1)
	w.stopChan = make(chan struct{})
	w.Requests = make(chan http.Request)
	w.logger = logger
	return nil
}

// Start : starts the server and waits for requests.
func (w *Webserver) Start() error {
	w.logger.Debugw(fmt.Sprintf("Webserver listening at %d!", w.port))
	w.wg.Add(1)
	// Start the server
	go func() {
		Server(w)
		// This will handle errors on its own
	}()

	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-w.stopChan:
				w.logger.Info("Closing...")
				return
			case r := <-w.Requests:
				go func() {
					responseChan := make(chan response2.Response)
					w.logger.Info("Received a photo ")
					name := r.FormValue("name")
					if name == "" {
						name = "temporary"
					}
					w.logger.Info("RECEIVED AN IMAGED NAMED  " + name)

					f, _, _ := r.FormFile("image")
					ctx := context.Background()

					w.Transactions <- transaction.InputTransaction{
						Transaction: transaction.Transaction{
							Payload: transaction.Payload{
								Reader:     io.Reader(f),
								ImageBytes: nil,
							},
							ImageData:    nil,
							ResponseChan: responseChan,
							Context:      ctx,
						},
						PipelineTag: "dummy",
					}
					response := <-responseChan
					w.logger.Info("RECEIVED RESPONSE.", zap.Any("response", response))
				}()
			}
		}
	}()
	return nil
}

//Close : graceful shutdown.
func (w *Webserver) Close() error {
	w.logger.Info("Sending closing signal...")
	//Gracefully closing the server.
	w.stopChan <- struct{}{}
	w.wg.Wait()
	err := w.Server.Shutdown(context.Background())
	if err != nil {
		return err
	}
	close(w.Transactions)
	w.logger.Info("Closed.")
	return nil
}
