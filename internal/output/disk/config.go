package disk

import (
	"os"

	cfg "github.com/sherifabdlnaby/prism/pkg/config"
)

//config struct
type config struct {
	Permission os.FileMode `mapstructure:"permission"`
	FilePath   string      `mapstructure:"filepath"`
	filepath   cfg.Selector
}

//defaultConfig returns the default configs
func defaultConfig() *config {
	return &config{
		Permission: 0777,
	}
}
