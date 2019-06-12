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
				Strategy: "embed",
				Pad:      "black",
			},
			Flip: flip{
				Direction: "none",
			},
			Scale: scale{
				Strategy: "embed",
				Pad:      "black",
			},
			//Label: label{
			//	Font:      "sans 10",
			//	Opacity:   "1",
			//	Alignment: "center",
			//	Color: rgb{
			//		R: "255",
			//		G: "255",
			//		B: "255",
			//	},
			//},
		},
	}
}

func DefaultTransformParams() *vips.TransformParams {
	return &vips.TransformParams{
		ResizeStrategy:          vips.ResizeStrategyEmbed,
		CropAnchor:              vips.AnchorAuto,
		ReductionSampler:        vips.KernelLanczos3,
		EnlargementInterpolator: vips.InterpolateBicubic,
	}
}
