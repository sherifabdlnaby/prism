package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	ResNotFound         = response{Code: http.StatusNotFound, Message: "not found"}
	ResMethodNotAllowed = response{Code: http.StatusMethodNotAllowed, Message: "method not allowed"}
	ResMissingFile      = response{Code: http.StatusBadRequest, Message: "not file uploaded, file name should be \"image\""}
	ResMissingPipeline  = response{Code: http.StatusBadRequest, Message: "pipeline field has dynamic values which are not present in the request"}
	ResNoAck            = response{Code: http.StatusBadRequest, Message: "request was dropped on purpose"}
	ResInternalError    = response{Code: http.StatusInternalServerError, Message: "internal server error"}
)

type response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func NewNoAck(noAck error) *response {
	return &response{http.StatusBadRequest, fmt.Sprintf("request was dropped, reason: %e", noAck)}
}

func NewError(err error) *response {
	return &response{http.StatusBadRequest, fmt.Sprintf("error while processing, reason: %e", err)}
}

func respondError(req *http.Request, w http.ResponseWriter, reply response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(reply.Code)
	jsonBuf, _ := json.Marshal(reply)
	_, _ = w.Write(jsonBuf)
}
