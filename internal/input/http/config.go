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

const (
	L_None = iota
	L_Debug
	L_Info
)

const (
	L_Success = iota + 1
	L_Fail
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
