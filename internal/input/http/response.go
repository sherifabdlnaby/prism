package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	errNotFound         = err{Code: http.StatusNotFound, Message: "not found"}
	errMethodNotAllowed = err{Code: http.StatusMethodNotAllowed, Message: "method not allowed"}
	errMissingFile      = err{Code: http.StatusBadRequest, Message: "not file uploaded, file name should be \"image\""}
	errMissingPipeline  = err{Code: http.StatusBadRequest, Message: "pipeline field has dynamic values which are not present in the request"}
	errInternalError    = err{Code: http.StatusInternalServerError, Message: "internal server error"}
)

type err struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func newNoAck(noAck error) *err {
	return &err{http.StatusBadRequest, fmt.Sprintf("request was dropped, reason: %s", noAck.Error())}
}

func newError(error error) *err {
	return &err{http.StatusBadRequest, fmt.Sprintf("error while processing, reason: %s", error.Error())}
}

func respondError(_ *http.Request, w http.ResponseWriter, error err) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(error.Code)
	jsonBuf, _ := json.Marshal(error)
	_, _ = w.Write(jsonBuf)
}

func respondMessage(_ *http.Request, w http.ResponseWriter, statusCode int, jsonMap interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	jsonBuf, _ := json.Marshal(jsonMap)
	_, _ = w.Write(jsonBuf)
}
