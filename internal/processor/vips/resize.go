package vips

import (
	"fmt"

	"github.com/sherifabdlnaby/bimg"
	cfg "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type resize struct {
	Raw       resizeRawConfig `mapstructure:",squash"`
	width     cfg.Selector
	height    cfg.Selector
	maxHeight cfg.Selector
	maxWidth  cfg.Selector
	minHeight cfg.Selector
	minWidth  cfg.Selector
	strategy  cfg.Selector
}

type resizeRawConfig struct {
	Width     string
	Height    string
	MaxHeight string `mapstructure:"max_height"`
	MaxWidth  string `mapstructure:"max_width"`
	MinHeight string `mapstructure:"min_height"`
	MinWidth  string `mapstructure:"min_width"`
	Strategy  string
}

func (o *resize) Init() (bool, error) {
	var err error

	if o.Raw == *resizeDefaults() {
		return false, nil
	}

	o.width, err = cfg.NewSelector(o.Raw.Width)
	if err != nil {
		return false, err
	}

	o.height, err = cfg.NewSelector(o.Raw.Height)
	if err != nil {
		return false, err
	}

	o.maxHeight, err = cfg.NewSelector(o.Raw.MaxHeight)
	if err != nil {
		return false, err
	}

	o.maxWidth, err = cfg.NewSelector(o.Raw.MaxWidth)
	if err != nil {
		return false, err
	}

	o.minHeight, err = cfg.NewSelector(o.Raw.MinHeight)
	if err != nil {
		return false, err
	}

	o.minWidth, err = cfg.NewSelector(o.Raw.MinWidth)
	if err != nil {
		return false, err
	}

	o.strategy, err = cfg.NewSelector(o.Raw.Strategy)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (o *resize) Apply(p *bimg.Options, data payload.Data) error {

	// --------------------------------------------------------------------

	width, err := o.width.EvaluateInt64(data)
	if err != nil {
		return err
	}

	height, err := o.height.EvaluateInt64(data)
	if err != nil {
		return err
	}

	maxWidth, err := o.maxWidth.EvaluateInt64(data)
	if err != nil {
		return err
	}

	maxHeight, err := o.maxHeight.EvaluateInt64(data)
	if err != nil {
		return err
	}

	minWidth, err := o.minWidth.EvaluateInt64(data)
	if err != nil {
		return err
	}

	minHeight, err := o.minHeight.EvaluateInt64(data)
	if err != nil {
		return err
	}

	strategy, err := o.strategy.Evaluate(data)
	if err != nil {
		return err
	}

	// --------------------------------------------------------------------

	p.Width = int(width)
	p.Height = int(height)
	p.MaxWidth = int(maxWidth)
	p.MaxHeight = int(maxHeight)
	p.MinWidth = int(minWidth)
	p.MinHeight = int(minHeight)
	p.Enlarge = true

	// --------------------------------------------------------------------

	switch strategy {
	case "embed":
		p.Embed = true
	case "crop":
		if p.Width > 0 && p.Height > 0 {
			p.Embed = true
			p.Crop = true
			break
		}
		p.Force = true
	case "stretch":
		p.Force = true
	default:
		err = fmt.Errorf("invalid value for field [strategy], got: %s", strategy)
	}

	return err
}
