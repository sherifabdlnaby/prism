package vips

import (
	"fmt"

	"github.com/h2non/bimg"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type crop struct {
	Raw    cropRawConfig `mapstructure:",squash"`
	width  config.Selector
	height config.Selector
	anchor config.Selector
}

type cropRawConfig struct {
	Width  string
	Height string
	Anchor string
}

func (o *crop) Init() (bool, error) {
	var err error

	if o.Raw == *cropDefaults() {
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

	o.anchor, err = config.NewSelector(o.Raw.Anchor)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (o *crop) Apply(p *bimg.Options, data payload.Data) error {
	var err error

	// // // // // // //

	width, err := o.width.EvaluateInt64(data)
	if err != nil {
		return nil
	}

	height, err := o.height.EvaluateInt64(data)
	if err != nil {
		return nil
	}

	// // // // // // //

	p.Width = int(width)
	p.Height = int(height)
	p.Crop = true

	anchor, err := o.anchor.Evaluate(data)
	if err != nil {
		return nil
	}

	switch anchor {
	case "center":
		p.Gravity = bimg.GravityCentre
	case "north":
		p.Gravity = bimg.GravityNorth
	case "east":
		p.Gravity = bimg.GravityEast
	case "south":
		p.Gravity = bimg.GravitySouth
	case "west":
		p.Gravity = bimg.GravityWest
	case "smart":
		p.Gravity = bimg.GravitySmart
	default:
		return fmt.Errorf("invalid value for field [anchor], got: %s", anchor)
	}

	return err
}
