package vips

import (
	"fmt"

	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type resize struct {
	Raw      resizeRawConfig `mapstructure:",squash"`
	width    config.Selector
	height   config.Selector
	strategy config.Selector
	pad      config.Selector
}

type resizeRawConfig struct {
	Width    string
	Height   string
	Strategy string
	Pad      string
}

func (o *resize) Init() (bool, error) {
	var err error

	if o.Raw == *resizeDefaults() {
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
	case "embed":
		p.ResizeStrategy = vips.ResizeStrategyEmbed
	case "crop":
		p.ResizeStrategy = vips.ResizeStrategyCrop
		p.CropAnchor = vips.AnchorCenter
	case "stretch":
		p.ResizeStrategy = vips.ResizeStrategyStretch
	default:
		err = fmt.Errorf("invalid value for field [strategy], got: %s", strategy)
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
