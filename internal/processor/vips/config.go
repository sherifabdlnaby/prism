package vips

import (
	"fmt"

	"github.com/sherifabdlnaby/govips/pkg/vips"
)

type Config struct {
	Operations Operations
	Export     Export

	export vips.ExportParams
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

func NewExportParams(e Export) (vips.ExportParams, error) {
	var ok bool
	var err error

	p := vips.ExportParams{}

	p.Format, ok = ImageTypeMap[e.Format]
	if !ok {
		return vips.ExportParams{}, fmt.Errorf("unsupported exporting file format")
	}

	p.Quality = e.Quality
	p.Compression = e.Compression
	p.Interlaced = e.Interlaced
	p.Lossless = e.Lossless
	p.StripProfile = e.StripProfile
	p.StripMetadata = e.StripMetadata
	p.BackgroundColor = &vips.Color{
		R: e.BackgroundColor.R,
		G: e.BackgroundColor.G,
		B: e.BackgroundColor.B,
	}

	p.Interpretation = vips.InterpretationSRGB

	return p, err
}

var ImageTypeMap = map[string]vips.ImageType{
	"jpeg": vips.ImageTypeJPEG,
	"jpg":  vips.ImageTypeJPEG,
	"png":  vips.ImageTypePNG,
	"tiff": vips.ImageTypeTIFF,
	"webp": vips.ImageTypeWEBP,
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
