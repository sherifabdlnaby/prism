package http

import cfg "github.com/sherifabdlnaby/prism/pkg/config"

type config struct {
	Port     int    `validate:"required"`
	FormName string `mapstructure:"form_name" validate:"required"`
	CertFile string
	KeyFile  string
	Paths    map[string]path `validate:"min=1"`
}

type path struct {
	Pipeline         string
	pipelineSelector cfg.Selector
}

func defaultConfig() *config {
	return &config{
		Port:     80,
		FormName: "image",
		CertFile: "",
		KeyFile:  "",
	}
}
