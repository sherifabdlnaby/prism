package http

import (
	"context"
	"fmt"
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"io"
	"net/http"
	"sync"
	"time"
)

// The Webserver receiving input.
//A lot of things to be added later: URL Signature key, HTTPs (key,cert), API key..etc
type Webserver struct {
	Transactions chan types.Transaction
	stopChan     chan struct{}
	logger       zap.Logger
	wg           sync.WaitGroup
	port         int
	CertFile     string
	KeyFile      string
	Server       *http.Server
	Requests     chan http.Request
}

//TransactionChan return transaction channel.
func (w *Webserver) TransactionChan() <-chan types.Transaction {
	return w.Transactions
}

//Init initializes webserver
func (w *Webserver) Init(config types.Config, logger zap.Logger) error {
	w.port = config["port"].(int)
	w.KeyFile = config["KeyFile"].(string)
	w.CertFile = config["CertFile"].(string)
	w.Transactions = make(chan types.Transaction, 1)
	w.stopChan = make(chan struct{})
	w.Requests = make(chan http.Request)
	w.logger = logger
	return nil
}

// Start : starts the server and waits for requests.
func (w *Webserver) Start() error {
	w.logger.Info(fmt.Sprintf("Webserver listening at %d!", w.port))
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
					responseChan := make(chan types.Response)
					w.logger.Info("Received a photo ")
					name := r.FormValue("name")
					if name == "" {
						name = "temporary"
					}
					w.logger.Info("RECEIVED AN IMAGED NAMED  " + name)

					f, _, _ := r.FormFile("image")

					w.Transactions <- types.Transaction{
						Payload: types.Payload{
							Name:      name,
							Reader:    io.Reader(f),
							ImageData: nil,
						},
						ResponseChan: responseChan,
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
func (w *Webserver) Close(time.Duration) error {
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
