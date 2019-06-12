package vips

import (
	"fmt"

	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type flip struct {
	Direction string
	direction config.Selector
}

func (o *flip) IsActive() bool {
	return o.direction.IsDynamic() || o.Direction != "none"
}

func (o *flip) Init() error {
	var err error

	o.direction, err = config.NewSelector(o.Direction)
	if err != nil {
		return err
	}

	return nil
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
