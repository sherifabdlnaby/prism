package nude

import "image/color"

type config struct {
	Drop   bool
	RGBA   rgba
	Export export
	rgba   color.RGBA
}

type export struct {
	Format  string `validate:"oneof=jpg jpeg png"`
	Quality int    `validate:"min=1,max=100"`
}

type rgba struct {
	R, G, B uint8
	A       float64 `validate:"min=0,max=1"`
}

//defaultConfig returns the default configs
func defaultConfig() *config {
	return &config{
		Drop: false,
		RGBA: rgba{
			R: 255,
			G: 255,
			B: 255,
			A: 1,
		},
		Export: export{
			Format:  "jpeg",
			Quality: 85,
		},
	}
}
