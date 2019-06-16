package vips

import (
	"github.com/sherifabdlnaby/bimg"
)

type Config struct {
	Operations Operations
	Export     export
}

type RGB struct {
	R, G, B uint8
}

func DefaultConfig() *Config {
	return &Config{
		Operations: Operations{
			Resize: resize{
				Raw: *resizeDefaults(),
			},
			Flip: flip{
				Raw: *flipDefaults(),
			},
			Blur: blur{
				Raw: *blurDefaults(),
			},
			Rotate: rotate{
				Raw: *rotateDefaults(),
			},
			Crop: crop{
				Raw: *cropDefaults(),
			},
			Label: label{
				Raw: *labelDefaults(),
			},
		},
		Export: export{
			Raw: exportRawConfig{
				Format:        "jpeg",
				Extend:        "black",
				Quality:       85,
				Compression:   6,
				StripMetadata: true,
			},
		},
	}
}

func resizeDefaults() *resizeRawConfig {
	return &resizeRawConfig{
		Width:    "",
		Height:   "",
		Strategy: "embed",
	}
}

func flipDefaults() *flipRawConfig {
	return &flipRawConfig{
		Direction: "",
	}
}

func blurDefaults() *blurRawConfig {
	return &blurRawConfig{
		Sigma: "",
	}
}

func rotateDefaults() *rotateRawConfig {
	return &rotateRawConfig{
		Angle: "",
	}
}

func labelDefaults() *labelRawConfig {
	return &labelRawConfig{
		Width:     "",
		DPI:       "10",
		Margin:    "0",
		Opacity:   "5",
		Replicate: "false",
		Text:      "",
		Font:      "sans 10",
		Color: rgb{
			R: "255",
			G: "255",
			B: "255",
		},
	}
}

func cropDefaults() *cropRawConfig {
	return &cropRawConfig{
		Width:  "",
		Height: "",
		Anchor: "center",
	}
}

// No need to use this, defaultOptions have it already embedded
func defaultExportOptions() *exportRawConfig {
	return &exportRawConfig{
		Format:        "jpeg",
		Quality:       85,
		Compression:   80,
		StripMetadata: false,
	}
}

func defaultOptions() *bimg.Options {
	return &bimg.Options{
		Quality:        85,
		Compression:    6,
		StripMetadata:  true,
		Type:           bimg.JPEG,
		Interpretation: bimg.InterpretationSRGB,
		Interlace:      true,
	}
}
