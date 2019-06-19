package http

import (
	"fmt"
	"net/http"

	"github.com/sherifabdlnaby/prism/pkg/payload"
	responseT "github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

// Server SETS THE CONFIGURATION AND STARTS THE SERVER.
func Server(w *Webserver) {
	addr := fmt.Sprintf(":%d", w.config.Port)

	// handler
	handler := ServerMux(w)

	w.Server = &http.Server{
		Addr:    addr,
		Handler: handler}

	err := Start(w)

	if err != nil {
		// TODO if server failed for some-reason
		//responseT handling should be implemented here.
	}
}

// Start MAY LOOK USELESS, BUT ITS HERE IN CASE WE ARE USING HTTPS.
func Start(w *Webserver) error {
	//Check if http has https files and then start https
	if w.config.CertFile != "" && w.config.KeyFile != "" {
		err := w.Server.ListenAndServeTLS(w.config.CertFile, w.config.KeyFile)
		return err
	}

	err := w.Server.ListenAndServe()
	return err
}

// ServerMux will have handlers.
func ServerMux(w *Webserver) http.Handler {
	mux := http.NewServeMux()

	//This is just an initial phase of the middleware.
	//Accept images at all the provided Paths.
	next := http.HandlerFunc(w.index)

	for path := range w.config.Paths {
		mux.Handle(path, next)
	}

	return mux
}

//Just an initial handler, receives a request, parses it to make sure its fine and sends a request to the plugin to be sent to processor.
func (w Webserver) index(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		if err := r.ParseMultipartForm(2 * 1024 * 1024); err != nil {
			respondError(r, rw, ResInternalError)
			return
		}

		err := r.ParseForm()

		file, _, err := r.FormFile("image")
		if err != nil {
			respondError(r, rw, ResMissingFile)
			return
		}

		//Setting all parameters
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
			respondError(r, rw, ResNotFound)
			return
		}

		pipeline, err := url.pipelineSelector.Evaluate(data)
		if err != nil {
			respondError(r, rw, ResMissingPipeline)
			return
		}

		responseChan := make(chan responseT.Response)
		w.Transactions <- transaction.InputTransaction{
			Transaction: transaction.Transaction{
				Payload:      file,
				Data:         data,
				Context:      r.Context(),
				ResponseChan: responseChan,
			},
			PipelineTag: pipeline,
		}

		response := <-responseChan

		if !response.Ack {
			// check if responseT is simply refused, or an internal responseT occured
			if response.AckErr != nil {
				respondError(r, rw, *NewNoAck(err))
			} else if response.Error != nil {
				respondError(r, rw, *NewError(err))
			}
			return
			//
		}

		rw.WriteHeader(http.StatusOK)

		return
	}

	rw.WriteHeader(http.StatusMethodNotAllowed)
}
