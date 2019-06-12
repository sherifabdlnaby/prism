package vips

import (
	"fmt"

	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type scale struct {
	Width    string
	Height   string
	Both     string
	Strategy string
	Pad      string

	width    config.Selector
	height   config.Selector
	both     config.Selector
	strategy config.Selector
	pad      config.Selector
}

func (o *scale) IsActive() bool {
	return o.Width != "" || o.Height != "" || o.Both != ""
}

func (o *scale) Init() error {
	var err error

	o.width, err = config.NewSelector(o.Width)
	if err != nil {
		return err
	}

	o.height, err = config.NewSelector(o.Height)
	if err != nil {
		return err
	}

	o.both, err = config.NewSelector(o.Both)
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

func (o *scale) Apply(p *vips.TransformParams, data payload.Data) error {

	// --------------------------------------------------------------------

	width, err := o.width.EvaluateFloat64(data)
	if err != nil {
		return err
	}

	height, err := o.height.EvaluateFloat64(data)
	if err != nil {
		return err
	}

	both, err := o.both.EvaluateFloat64(data)
	if err != nil {
		return err
	}

	p.Width.SetScale(width)
	p.Height.SetScale(height)

	if both != 0.0 {
		p.Height.SetScale(both)
		p.Width.SetScale(both)
	}

	// --------------------------------------------------------------------

	strategy, err := o.strategy.Evaluate(data)
	if err != nil {
		return err
	}

	switch strategy {
	case "embed":
		p.ResizeStrategy = vips.ResizeStrategyEmbed
	case "crop":
		p.ResizeStrategy = vips.ResizeStrategyCrop
	case "stretch":
		p.ResizeStrategy = vips.ResizeStrategyStretch
	default:
		err = fmt.Errorf("invalid value for field [strategy], got: %s", strategy)
	}

	// --------------------------------------------------------------------

	pad, err := o.pad.Evaluate(data)
	if err != nil {
		return err
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
	default:
		err = fmt.Errorf("invalid value for field [pad], got: %s", pad)
	}

	return err
}
