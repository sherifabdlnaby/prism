package http

import (
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	responseT "github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"
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

//Request Logger logs every request
func requestLogger(next http.Handler, lType int, l zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addr := r.RemoteAddr
		if i := strings.LastIndex(addr, ":"); i != -1 {
			addr = addr[:i]
		}
		switch lType {
		case LDebug:
			l.Debugw("RECIEVED A REQUEST",
				"IP", addr,
				"TIME", time.Now().Format("02/Jan/2006:15:04:05"),
				"FORM", fmt.Sprintf("%s %s %s", r.Method, r.URL, r.Proto),
				"USERAGENT", r.UserAgent())
		case LInfo:
			l.Infow("RECIEVED A REQUEST",
				"IP", addr,
				"TIME", time.Now().Format("02/Jan/2006:15:04:05"),
				"FORM", fmt.Sprintf("%s %s %s", r.Method, r.URL, r.Proto),
				"USERAGENT", r.UserAgent())
		}
		next.ServeHTTP(w, r)
	})
}

//Ratelimiter will rate limit requests to avoid denial of service attacks.
func rateLimiter(next http.Handler, l float64, ws *Webserver) http.Handler {
	lmt := tollbooth.NewLimiter(l, nil)
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		respondError(r, w, resRateLimit, ws)
		return
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

	if w.config.LogRequest != LNone {
		next = http.Handler(requestLogger(next, w.config.LogRequest, w.logger))
	}

	if w.config.RateLimit > 0 {
		next = http.Handler(rateLimiter(next, w.config.RateLimit, w))
	}

	for path := range w.config.Paths {
		mux.Handle(path, next)
	}

	return mux
}

//handle will formulate request into a transaction and await response
func (w *Webserver) handle(rw http.ResponseWriter, r *http.Request) {
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
			respondError(r, rw, resNotFound, w)
			return
		}

		pipeline, err := url.pipelineSelector.Evaluate(data)
		if err != nil {
			respondError(r, rw, resMissingPipeline, w)
			return
		}

		// Multi-reader
		reader, err := r.MultipartReader()
		if err != nil {
			respondError(r, rw, *newError(err), w)
			return
		}

		// Get a part from multipart
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				respondError(r, rw, resMissingFile, w)
			} else {
				respondError(r, rw, *newError(err), w)
			}
			return
		}

		// Check is part form name is as configured
		if part.FormName() != w.config.FormName {
			respondError(r, rw, resMissingFile, w)
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
				respondError(r, rw, *newNoAck(response.AckErr), w)
			} else if response.Error != nil {
				respondError(r, rw, *newError(response.Error), w)
			}
			return

		}

		if w.config.LogResponse == LSuccess {
			w.logger.Debugw("Successful request ",
				"TIME", time.Now().Format("02/Jan/2006:15:04:05 -0700"),
				"FORM", fmt.Sprintf("%s %s %s", r.Method, r.URL, r.Proto))
		}
		rw.WriteHeader(http.StatusOK)

		return
	}

	respondError(r, rw, resMethodNotAllowed, w)
}
