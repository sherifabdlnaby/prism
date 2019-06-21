package http

import (
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ENUM for logging requests
// LFail:	Log all requests via Logger debug
// LogInfo:   Log all requests via logger info
const (
	LogDebug = iota
	LogInfo
)

// Used to capture status in logging
type logResponseWriter struct {
	http.ResponseWriter
	status int
}

func (rec *logResponseWriter) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

//Request Logger logs every request
func requestLogger(next http.Handler, lType int, l zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logWriter := &logResponseWriter{
			ResponseWriter: w,
		}

		start := time.Now()
		next.ServeHTTP(logWriter, r)
		finish := time.Since(start)

		addr := r.RemoteAddr

		if i := strings.LastIndex(addr, ":"); i != -1 {
			addr = addr[:i]
		}

		switch lType {
		case LogDebug:
			l.Debugf("request \t %v \t %s \t %d \t %s \t %s \t %s \t %s \t %s", finish, r.Method, logWriter.status, r.Proto, r.URL.Path, r.URL.RawQuery, r.UserAgent(), addr)
		case LogInfo:
			l.Infow("REQUEST",
				"ELAPSED", finish,
				"METHOD", r.Method,
				"STATUS", logWriter.status,
				"PROTO", r.Proto,
				"PATH", r.URL.Path,
				"QUERY", r.URL.RawQuery,
				"USERAGENT", r.UserAgent(),
				"IP", addr,
			)
		}
	})
}
