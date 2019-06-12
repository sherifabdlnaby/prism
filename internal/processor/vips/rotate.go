package vips

import (
	"fmt"

	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type rotate struct {
	Angle string
	angle config.Selector
}

func (o *rotate) IsActive() bool {
	return o.angle.IsDynamic() || o.Angle != ""
}

func (o *rotate) Init() error {
	var err error

	o.angle, err = config.NewSelector(o.Angle)
	if err != nil {
		return err
	}

	return nil
}

func (o *rotate) Apply(p *vips.TransformParams, data payload.Data) error {

	angle, err := o.angle.Evaluate(data)

	if err != nil {
		return err
	}

	switch angle {
	case "0":
		p.Rotate = vips.Angle0
	case "90":
		p.Rotate = vips.Angle90
	case "180":
		p.Rotate = vips.Angle180
	case "270":
		p.Rotate = vips.Angle270
	default:
		err = fmt.Errorf("invalid value for field [angle], got: %s", angle)
	}

	return err
}
