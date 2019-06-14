package vips

import (
	"github.com/h2non/bimg"
)

type Config struct {
	Operations Operations
	Export     Export
}

type Export struct {
	Format          string
	Quality         int
	Compression     int
	Interlaced      bool
	Lossless        bool
	StripProfile    bool
	StripMetadata   bool
	BackgroundColor RGB
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
			Scale: scale{
				Raw: *scaleDefaults(),
			},
			Crop: crop{
				Raw: *cropDefaults(),
			},
			Label: label{
				Raw: *labelDefaults(),
			},
		},
		Export: *DefaultExport(),
	}
}

func DefaultExport() *Export {
	return &Export{
		Format:        "jpeg",
		Quality:       90,
		Compression:   0,
		StripMetadata: true,
		BackgroundColor: RGB{
			R: 255,
			G: 255,
			B: 255,
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

func scaleDefaults() *scaleRawConfig {
	return &scaleRawConfig{
		Width:    "",
		Height:   "",
		Both:     "",
		Strategy: "embed",
		Pad:      "black",
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

func defaultOptions() *bimg.Options {
	return &bimg.Options{}
}
