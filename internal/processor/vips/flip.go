package vips

import (
	"fmt"

	"github.com/h2non/bimg"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type flip struct {
	Raw       flipRawConfig `mapstructure:",squash"`
	direction config.Selector
}

type flipRawConfig struct {
	Direction string
}

func (o *flip) Init() (bool, error) {
	var err error

	if o.Raw == *flipDefaults() {
		return false, nil
	}

	o.direction, err = config.NewSelector(o.Raw.Direction)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (o *flip) Apply(p *bimg.Options, data payload.Data) error {

	// --------------------------------------------------------------------

	direction, err := o.direction.Evaluate(data)
	if err != nil {
		return err
	}

	// --------------------------------------------------------------------

	switch direction {
	case "horizontal":
		p.Flip = true
	case "vertical":
		p.Flop = true
	case "both":
		p.Flip = true
		p.Flop = true
	case "none":
	default:
		err = fmt.Errorf("invalid value for field [direction], got: %s", direction)
	}

	return err
}
