package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	resNotFound         = response{Code: http.StatusNotFound, Message: "not found"}
	resMethodNotAllowed = response{Code: http.StatusMethodNotAllowed, Message: "method not allowed"}
	resMissingFile      = response{Code: http.StatusBadRequest, Message: "not file uploaded, file name should be \"image\""}
	resMissingPipeline  = response{Code: http.StatusBadRequest, Message: "pipeline field has dynamic values which are not present in the request"}
	resNoAck            = response{Code: http.StatusBadRequest, Message: "request was dropped on purpose"}
	resInternalError    = response{Code: http.StatusInternalServerError, Message: "internal server error"}
	resRateLimit        = response{Code: http.StatusTooManyRequests, Message: "Too many requests"}
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

func respondError(r *http.Request, w http.ResponseWriter, reply response, ws *Webserver) {
	if ws.config.LogResponse == L_Fail {
		addr := r.RemoteAddr
		if i := strings.LastIndex(addr, ":"); i != -1 {
			addr = addr[:i]
		}
		ws.logger.Debugw(reply.Message,
			"IP", addr,
			"TIME", time.Now().Format("02/Jan/2006:15:04:05"))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(reply.Code)

	jsonBuf, _ := json.Marshal(reply)
	_, _ = w.Write(jsonBuf)
}
