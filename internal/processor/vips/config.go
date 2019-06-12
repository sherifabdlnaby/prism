package vips

import "github.com/sherifabdlnaby/govips/pkg/vips"

type Config struct {
	Operations Operations
}

func DefaultConfig() *Config {
	return &Config{
		Operations{
			Resize: resize{
				Strategy: "auto",
				Pad:      "black",
			},
			Flip: flip{
				Direction: "none",
			},
		},
	}
}

func DefaultTransformParams() *vips.TransformParams {
	return &vips.TransformParams{
		ResizeStrategy:          vips.ResizeStrategyAuto,
		CropAnchor:              vips.AnchorAuto,
		ReductionSampler:        vips.KernelLanczos3,
		EnlargementInterpolator: vips.InterpolateBicubic,
	}
}
