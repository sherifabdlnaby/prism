package vips

import (
	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type resize struct {
	Width    string
	Height   string
	Strategy string `validate:"oneof=auto embed crop stretch"`
	Pad      string `validate:"oneof=black copy repeat mirror white background"`

	width    config.Selector
	height   config.Selector
	strategy config.Selector
	pad      config.Selector
}

func (o *resize) Init() error {
	var err error

	o.width, err = config.NewSelector(o.Width)
	if err != nil {
		return err
	}

	o.height, err = config.NewSelector(o.Height)
	if err != nil {
		return err
	}

	o.strategy, err = config.NewSelector(o.Strategy)
	if err != nil {
		return err
	}

	o.pad, err = config.NewSelector(o.Pad)
	if err != nil {
		return err
	}

	return nil
}

func (o *resize) Apply(p *vips.TransformParams, data payload.Data) error {

	// --------------------------------------------------------------------

	width, err := o.width.EvaluateInt64(data)
	if err != nil {
		return err
	}

	height, err := o.height.EvaluateInt64(data)
	if err != nil {
		return err
	}

	strategy, err := o.strategy.Evaluate(data)
	if err != nil {
		return err
	}

	pad, err := o.pad.Evaluate(data)
	if err != nil {
		return err
	}

	// --------------------------------------------------------------------

	p.Width.SetInt(int(width))

	p.Height.SetInt(int(height))

	// --------------------------------------------------------------------

	switch strategy {
	case "auto":
		p.ResizeStrategy = vips.ResizeStrategyAuto
	case "embed":
		p.ResizeStrategy = vips.ResizeStrategyEmbed
	case "crop":
		p.ResizeStrategy = vips.ResizeStrategyCrop
	case "stretch":
		p.ResizeStrategy = vips.ResizeStrategyStretch
	}

	switch pad {
	case "black":
		p.PadStrategy = vips.ExtendBlack
	case "copy":
		p.PadStrategy = vips.ExtendCopy
	case "repeat":
		p.PadStrategy = vips.ExtendRepeat
	case "mirror":
		p.PadStrategy = vips.ExtendMirror
	case "white":
		p.PadStrategy = vips.ExtendWhite
	case "background":
		p.PadStrategy = vips.ExtendBackground

	}

	return nil
}
