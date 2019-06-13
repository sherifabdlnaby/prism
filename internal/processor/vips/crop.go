package vips

import (
	"fmt"

	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type crop struct {
	Width  string
	Height string
	Anchor string

	width  config.Selector
	height config.Selector
	anchor config.Selector
}

func (o *crop) IsActive() bool {
	return o.anchor.IsDynamic() || o.Anchor != ""
}

func (o *crop) Init() error {
	var err error

	o.width, err = config.NewSelector(o.Width)
	if err != nil {
		return err
	}
	o.height, err = config.NewSelector(o.Height)
	if err != nil {
		return err
	}
	o.anchor, err = config.NewSelector(o.Anchor)
	if err != nil {
		return err
	}

	return nil
}

func (o *crop) Apply(p *vips.TransformParams, data payload.Data) error {
	var err error

	//enable croping
	p.ResizeStrategy = vips.ResizeStrategyCrop

	// // // // // // //

	width, err := o.width.EvaluateInt64(data)
	if err != nil {
		return nil
	}
	p.Width.SetInt(int(width))

	height, err := o.height.EvaluateInt64(data)
	if err != nil {
		return nil
	}
	p.Height.SetInt(int(height))

	// // // // // // //

	anchor, err := o.anchor.Evaluate(data)
	if err != nil {
		return nil
	}

	switch anchor {
	case "center":
		p.CropAnchor = vips.AnchorCenter
	case "entropy":
		p.CropAnchor = vips.AnchorEntropy
	case "face":
		p.CropAnchor = vips.AnchorFace
	case "bottom":
		p.CropAnchor = vips.AnchorBottom
	case "bottom_left":
		p.CropAnchor = vips.AnchorBottomLeft
	case "bottom_right":
		p.CropAnchor = vips.AnchorBottomRight
	case "left":
		p.CropAnchor = vips.AnchorLeft
	case "right":
		p.CropAnchor = vips.AnchorRight
	case "top":
		p.CropAnchor = vips.AnchorTop
	case "top_left":
		p.CropAnchor = vips.AnchorTopLeft
	case "top_right":
		p.CropAnchor = vips.AnchorTopRight
	default:
		return fmt.Errorf("invalid value for field [anchor], got: %s", anchor)
	}

	return err
}
