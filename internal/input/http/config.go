package http

import cfg "github.com/sherifabdlnaby/prism/pkg/config"

type config struct {
	Port       int    `validate:"required"`
	ImageField string `mapstructure:"image_field" validate:"required"`
	CertFile   string
	KeyFile    string
	Paths      map[string]path `validate:"min=1"`
	LogRequest string          `mapstructure:"log_request" validate:"oneof=all debug none"`
	LogErrors  bool            `mapstructure:"log_errors"`
	RateLimit  float64         `mapstructure:"rate_limit"`
}

type path struct {
	Pipeline         string
	pipelineSelector cfg.Selector
}

func defaultConfig() *config {
	return &config{
		Port:       80,
		ImageField: "image",
		CertFile:   "",
		KeyFile:    "",
		LogRequest: "all",
		LogErrors:  false,
		RateLimit:  20,
	}
}
