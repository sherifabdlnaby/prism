package vips

import (
	"github.com/sherifabdlnaby/govips/pkg/vips"
)

type Config struct {
	Operations Operations
}

func DefaultConfig() *Config {
	return &Config{
		Operations{
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
		},
	}
}

func resizeDefaults() *resizeRawConfig {
	return &resizeRawConfig{
		Width:    "",
		Height:   "",
		Strategy: "embed",
		Pad:      "black",
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

func cropDefaults() *cropRawConfig {
	return &cropRawConfig{
		Width:  "",
		Height: "",
		Anchor: "center",
	}
}

func defaultTransformParams() *vips.TransformParams {
	return &vips.TransformParams{
		ResizeStrategy:          vips.ResizeStrategyEmbed,
		CropAnchor:              vips.AnchorAuto,
		ReductionSampler:        vips.KernelLanczos3,
		EnlargementInterpolator: vips.InterpolateBicubic,
	}
}
