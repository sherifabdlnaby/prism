package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sherifabdlnaby/prism/pkg/payload"
	responseT "github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

// buildServer build handlers and server
func (w *Webserver) buildServer() {
	addr := fmt.Sprintf(":%d", w.config.Port)

	// create handlers
	handler := buildHandlers(w)

	w.Server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}

// listenAndServe start listening
func (w *Webserver) listenAndServe() error {
	//Check if http has https files and then start https
	// TODO check if this is working
	if w.config.CertFile != "" && w.config.KeyFile != "" {
		err := w.Server.ListenAndServeTLS(w.config.CertFile, w.config.KeyFile)
		return err
	}

	err := w.Server.ListenAndServe()
	return err
}

// buildHandlers will build handlers according to config
func buildHandlers(w *Webserver) http.Handler {
	mux := http.NewServeMux()

	//This is just an initial phase of the middleware.
	//Accept images at all the provided Paths.
	next := http.HandlerFunc(w.handle)

	for path := range w.config.Paths {
		mux.Handle(path, next)
	}

	return mux
}

//handle will formulate request into a transaction and await response
func (w Webserver) handle(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		// Parse form into Data
		err := r.ParseForm()
		data := make(payload.Data, len(r.Form))
		for key := range r.Form {
			if len(r.Form[key]) > 1 {
				data[key] = r.Form[key]
				continue
			}
			data[key] = r.Form[key][0]
		}

		url, ok := w.config.Paths[r.URL.Path]
		if !ok {
			respondError(r, rw, resNotFound)
			return
		}

		pipeline, err := url.pipelineSelector.Evaluate(data)
		if err != nil {
			respondError(r, rw, resMissingPipeline)
			return
		}

		// Multi-reader
		reader, err := r.MultipartReader()
		if err != nil {
			respondError(r, rw, *newError(err))
			return
		}

		// Get a part from multipart
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				respondError(r, rw, resMissingFile)
			} else {
				respondError(r, rw, *newError(err))
			}
			return
		}

		// Check is part form name is as configured
		if part.FormName() != w.config.FormName {
			respondError(r, rw, resMissingFile)
			return
		}

		responseChan := make(chan responseT.Response)
		w.Transactions <- transaction.InputTransaction{
			Transaction: transaction.Transaction{
				Payload:      payload.Stream(part),
				Data:         data,
				Context:      r.Context(),
				ResponseChan: responseChan,
			},
			PipelineTag: pipeline,
		}

		// Wait Response
		response := <-responseChan

		if !response.Ack {
			// check if responseT is simply refused, or an internal responseT occured
			if response.AckErr != nil {
				respondError(r, rw, *newNoAck(response.AckErr))
			} else if response.Error != nil {
				respondError(r, rw, *newError(response.Error))
			}
			return
			//
		}

		rw.WriteHeader(http.StatusOK)

		return
	}

	respondError(r, rw, resMethodNotAllowed)
}
