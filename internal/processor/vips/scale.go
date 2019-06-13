package vips

import (
	"fmt"

	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type scale struct {
	Raw      scaleRawConfig `mapstructure:",squash"`
	width    config.Selector
	height   config.Selector
	both     config.Selector
	strategy config.Selector
	pad      config.Selector
}

type scaleRawConfig struct {
	Width    string
	Height   string
	Both     string
	Strategy string
	Pad      string
}

func (o *scale) Init() (bool, error) {
	var err error

	if o.Raw == *scaleDefaults() {
		return false, nil
	}

	o.width, err = config.NewSelector(o.Raw.Width)
	if err != nil {
		return false, err
	}

	o.height, err = config.NewSelector(o.Raw.Height)
	if err != nil {
		return false, err
	}

	o.both, err = config.NewSelector(o.Raw.Both)
	if err != nil {
		return false, err
	}

	o.strategy, err = config.NewSelector(o.Raw.Strategy)
	if err != nil {
		return false, err
	}

	o.pad, err = config.NewSelector(o.Raw.Pad)
	if err != nil {
		return false, err
	}

	return true, nil
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
