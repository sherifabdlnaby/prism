package vips

import (
	"fmt"

	"github.com/sherifabdlnaby/govips/pkg/vips"
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

func (o *flip) Apply(p *vips.TransformParams, data payload.Data) error {

	// --------------------------------------------------------------------

	direction, err := o.direction.Evaluate(data)
	if err != nil {
		return err
	}

	// --------------------------------------------------------------------

	switch direction {
	case "horizontal":
		p.Flip = vips.FlipHorizontal
	case "vertical":
		p.Flip = vips.FlipVertical
	case "both":
		p.Flip = vips.FlipBoth
	case "none":
		p.Flip = vips.FlipNone
	default:
		err = fmt.Errorf("invalid value for field [direction], got: %s", direction)
	}

	return err
}
