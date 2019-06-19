package http

import cfg "github.com/sherifabdlnaby/prism/pkg/config"

type config struct {
	Port       int    `validate:"required"`
	ImageField string `mapstructure:"image_field" validate:"required"`
	CertFile   string
	KeyFile    string
	Paths      map[string]path `validate:"min=1"`
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
	}
}
