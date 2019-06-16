package vips

import (
	"fmt"

	"github.com/sherifabdlnaby/bimg"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type rotate struct {
	Raw   rotateRawConfig `mapstructure:",squash"`
	angle config.Selector
}

type rotateRawConfig struct {
	Angle string
}

func (o *rotate) Init() (bool, error) {
	var err error

	if o.Raw == *rotateDefaults() {
		return false, nil
	}

	o.angle, err = config.NewSelector(o.Raw.Angle)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (o *rotate) Apply(p *bimg.Options, data payload.Data) error {

	angle, err := o.angle.Evaluate(data)

	if err != nil {
		return err
	}

	switch angle {
	case "0":
		p.Rotate = bimg.D0
	case "90":
		p.Rotate = bimg.D90
	case "180":
		p.Rotate = bimg.D180
	case "270":
		p.Rotate = bimg.D270
	default:
		err = fmt.Errorf("invalid value for field [angle], got: %s", angle)
	}

	return err
}
