package http

import cfg "github.com/sherifabdlnaby/prism/pkg/config"

type config struct {
	Port        int    `validate:"required"`
	FormName    string `mapstructure:"form_name" validate:"required"`
	CertFile    string
	KeyFile     string
	Paths       map[string]path `validate:"min=1"`
	LogRequest  int
	LogResponse int
	RateLimit   float64
}

type path struct {
	Pipeline         string
	pipelineSelector cfg.Selector
}

// ENUM for logging requests
// LNone: doesn't log any request
// LFail:	Log all requests via Logger debug
// LInfo: Log all requests via logger info
const (
	LNone = iota
	LDebug
	LInfo
)

// ENUM for logging responses
// LSuccess: logging successful requests
// LFail:	logging failed	requests
const (
	LSuccess = iota + 1
	LFail
)

func defaultConfig() *config {
	return &config{
		Port:        80,
		FormName:    "image",
		CertFile:    "",
		KeyFile:     "",
		LogRequest:  0,
		LogResponse: 0,
		RateLimit:   0,
	}
}
