package dummy

import cfg "github.com/sherifabdlnaby/prism/pkg/config"

// config struct used to decode YAML into
type config struct {
	FileName string
	Pipeline string
	Tick     int
	Timeout  int
	Count    int

	pipeline cfg.Selector
	filename cfg.Selector
}

func defaultConfig() *config {
	return &config{
		Tick: 1000,
	}
}
