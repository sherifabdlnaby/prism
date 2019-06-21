package http

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/didip/tollbooth"
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
	if w.config.CertFile != "" && w.config.KeyFile != "" {
		err := w.Server.ListenAndServeTLS(w.config.CertFile, w.config.KeyFile)
		return err
	}

	err := w.Server.ListenAndServe()
	return err
}

//Ratelimiter will rate limit requests to avoid denial of service attacks.
func rateLimiter(next http.Handler, reqPerSecond float64, ws *Webserver) http.Handler {
	lmt := tollbooth.NewLimiter(reqPerSecond, nil)
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		ws.respondError(r, w, resRateLimit)
	})
	middle := func(w http.ResponseWriter, r *http.Request) {
		httpError := tollbooth.LimitByRequest(lmt, w, r)
		if httpError != nil {
			lmt.ExecOnLimitReached(w, r)
			return
		}
		// There's no rate-limit error, serve the next handler.
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(middle)
}

// buildHandlers will build handlers according to config
func buildHandlers(w *Webserver) http.Handler {
	mux := http.NewServeMux()

	//This is just an initial phase of the middleware.
	//Accept images at all the provided Paths.
	next := http.Handler(http.HandlerFunc(w.handle))

	if w.config.LogRequest != "none" {
		if w.config.LogRequest == "all" {
			next = http.Handler(requestLogger(next, LogInfo, w.logger))
		}
		if w.config.LogRequest == "debug" {
			next = http.Handler(requestLogger(next, LogDebug, w.logger))
		}
	}

	if w.config.RateLimit > 0 {
		next = http.Handler(rateLimiter(next, w.config.RateLimit, w))
	}

	for path := range w.config.Paths {
		mux.Handle(path, next)
	}

	//register index
	mux.HandleFunc("/index", w.index)

	return mux
}

//handle will formulate request into a transaction and await err
func (w *Webserver) handle(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		// Parse form into Data
		err := r.ParseForm()
		if err != nil {
			w.respondError(r, rw, errMissingPipeline)
			return
		}

		data := make(payload.Data, len(r.Form))
		for key := range r.Form {
			if len(r.Form[key]) > 1 {
				data[key] = r.Form[key]
				continue
			}
			data[key] = r.Form[key][0]
		}

		path, ok := w.config.Paths[r.URL.Path]
		if !ok {
			w.respondError(r, rw, errNotFound)
			return
		}

		pipeline, err := path.pipelineSelector.Evaluate(data)
		if err != nil {
			w.respondError(r, rw, errMissingPipeline)
			return
		}

		// Multi-reader
		reader, err := r.MultipartReader()
		if err != nil {
			w.respondError(r, rw, *newError(err))
			return
		}

		// Get a part from multipart
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				w.respondError(r, rw, errMissingFile)
			} else {
				w.respondError(r, rw, *newError(err))
			}
			return
		}

		// Check is part form name is as configured
		if part.FormName() != w.config.ImageField {
			w.respondError(r, rw, errMissingFile)
			return
		}

		// Add filename to Data (and remove extension
		filename := part.FileName()
		data["_filename"] = filename[0 : len(filename)-len(filepath.Ext(filename))]

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
			// check if responseT is simply refused, or an internal responseT occurred
			if response.AckErr != nil {
				w.respondError(r, rw, *newNoAck(response.AckErr))
			} else if response.Error != nil {
				w.respondError(r, rw, *newError(response.Error))
			}
			return

		}

		w.respondMessage(r, rw, resSuccess)

		return
	} else if r.Method == http.MethodGet {
		w.respondJSON(r, rw, http.StatusOK, map[string]interface{}{
			"message":  "Prism HTTP Server, use POST multipart/form-data requests on this path.",
			"pipeline": w.config.Paths[r.URL.Path].Pipeline,
			"version":  version,
		})
		return
	}

	w.respondError(r, rw, errMethodNotAllowed)
}

//handle will formulate request into a transaction and await err
func (w *Webserver) index(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.respondJSON(r, rw, http.StatusOK, map[string]interface{}{
			"message":          "Prism HTTP Server",
			"version":          version,
			"registered_paths": len(w.config.Paths),
		})
		return
	}
	w.respondError(r, rw, errMethodNotAllowed)
}
