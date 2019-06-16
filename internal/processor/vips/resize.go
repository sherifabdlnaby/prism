package vips

import (
	"fmt"

	"github.com/sherifabdlnaby/bimg"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type resize struct {
	Raw      resizeRawConfig `mapstructure:",squash"`
	width    config.Selector
	height   config.Selector
	strategy config.Selector
}

type resizeRawConfig struct {
	Width    string
	Height   string
	Strategy string
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

	strategy, err := o.strategy.Evaluate(data)
	if err != nil {
		return err
	}

	// --------------------------------------------------------------------

	p.Width = int(width)
	p.Height = int(height)

	// --------------------------------------------------------------------

	switch strategy {
	case "embed":
		p.Embed = true
	case "crop":
		p.Embed = true
		p.Crop = true
	case "stretch":
		p.Force = true
	default:
		err = fmt.Errorf("invalid value for field [strategy], got: %s", strategy)
	}

	return err
}
