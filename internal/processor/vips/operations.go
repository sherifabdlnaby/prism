package vips

import (
	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type Operations struct {
	Resize resize
	Flip   flip
}

func (o *Operations) Init() error {
	err := o.Resize.Init()
	if err != nil {
		return err
	}

	err = o.Flip.Init()
	if err != nil {
		return err
	}

	return err
}

func (o *Operations) Do(image *vips.ImageRef, data payload.Data) error {
	var err error
	params := &vips.TransformParams{
		ResizeStrategy:          vips.ResizeStrategyAuto,
		CropAnchor:              vips.AnchorAuto,
		ReductionSampler:        vips.KernelLanczos3,
		EnlargementInterpolator: vips.InterpolateBicubic,
	}

	// check resize
	err = o.Resize.Apply(params, data)
	if err != nil {
		return err
	}

	err = o.Flip.Apply(params, data)
	if err != nil {
		return err
	}

	bb := vips.NewBlackboard(image.Image(), image.Format(), params)
	err = vips.ProcessBlackboard(bb)
	image.SetImage(bb.Image())

	if err != nil {
		return err
	}

	return nil
}
