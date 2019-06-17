package vips

import (
	"github.com/sherifabdlnaby/bimg"
)

// config struct used to decode YAML into
type config struct {
	Operations operations
	Export     export
}

// defaultConfig return default configuration for VIPS plugin configuration
func defaultConfig() *config {
	return &config{
		Operations: operations{
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
			Raw: *exportDefaults(),
		},
	}
}

// resizeDefaults return default configuration for resize operation configuration
func resizeDefaults() *resizeRawConfig {
	return &resizeRawConfig{
		Width:    "",
		Height:   "",
		Strategy: "embed",
	}
}

// flipDefaults return default configuration for flip operation configuration
func flipDefaults() *flipRawConfig {
	return &flipRawConfig{
		Direction: "",
	}
}

// blurDefaults return default configuration for blur operation configuration
func blurDefaults() *blurRawConfig {
	return &blurRawConfig{
		Sigma: "",
	}
}

// rotateDefaults return default configuration for rotate operation configuration
func rotateDefaults() *rotateRawConfig {
	return &rotateRawConfig{
		Angle: "",
	}
}

// labelDefaults return default configuration for label operation configuration
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

// cropDefaults return default configuration for crop operation configuration
func cropDefaults() *cropRawConfig {
	return &cropRawConfig{
		Width:  "",
		Height: "",
		Anchor: "center",
	}
}

// exportDefaults return default exporting configuration for VIPS plugin
func exportDefaults() *exportRawConfig {
	return &exportRawConfig{
		Format:        "jpeg",
		Extend:        "black",
		Quality:       85,
		Compression:   6,
		StripMetadata: true,
	}
}

// defaultOptions return default bimg.Options for internal usage
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
