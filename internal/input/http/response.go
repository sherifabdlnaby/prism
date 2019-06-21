package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	resNotFound         = response{Code: http.StatusNotFound, Message: "not found"}
	resMethodNotAllowed = response{Code: http.StatusMethodNotAllowed, Message: "method not allowed"}
	resMissingFile      = response{Code: http.StatusBadRequest, Message: "not file uploaded, file name should be \"image\""}
	resMissingPipeline  = response{Code: http.StatusBadRequest, Message: "pipeline field has dynamic values which are not present in the request"}
	resNoAck            = response{Code: http.StatusBadRequest, Message: "request was dropped on purpose"}
	resInternalError    = response{Code: http.StatusInternalServerError, Message: "internal server error"}
	resRateLimit        = response{Code: http.StatusTooManyRequests, Message: "Too many requests"}
	resSuccess          = response{Code: http.StatusOK, Message: "Request Successful"}
)

type response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func newNoAck(noAck error) *response {
	return &response{http.StatusBadRequest, fmt.Sprintf("request was dropped, reason: %s", noAck.Error())}
}

func newError(err error) *response {
	return &response{http.StatusBadRequest, fmt.Sprintf("error while processing, reason: %s", err.Error())}
}

func (w *Webserver) respondError(r *http.Request, wr http.ResponseWriter, reply response) {
	if w.config.LogErrors {
		w.logger.Error(reply.Message)
	}

	wr.Header().Set("Content-Type", "application/json")
	wr.WriteHeader(reply.Code)

	jsonBuf, _ := json.Marshal(reply)
	_, _ = wr.Write(jsonBuf)
}

func (w *Webserver) respondMessage(r *http.Request, rw http.ResponseWriter, reply response) {

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(reply.Code)

	jsonBuf, _ := json.Marshal(reply)
	_, _ = rw.Write(jsonBuf)
}
