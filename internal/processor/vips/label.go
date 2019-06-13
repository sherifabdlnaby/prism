package vips

import (
	"fmt"

	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

// TODO : This is NOT Used because on PNG images it return an error
// ifthenelse: not one band or 4 bands
// This is a problem with govips, fix that to enable this op
type label struct {
	Text           string
	Font           string
	Width          string
	Height         string
	RelativeDim    string `mapstructure:"relative_dim"`
	OffsetX        string
	OffsetY        string
	RelativeOffset string `mapstructure:"relative_offset"`
	Opacity        string
	Color          rgb
	Alignment      string

	text           config.Selector
	font           config.Selector
	width          config.Selector
	height         config.Selector
	relativeDim    config.Selector
	offsetX        config.Selector
	offsetY        config.Selector
	relativeOffset config.Selector
	opacity        config.Selector
	colorR         config.Selector
	colorG         config.Selector
	colorB         config.Selector
	alignment      config.Selector
}

type rgb struct {
	R, G, B string
}

func (o *label) Init() (bool, error) {
	var err error

	o.text, err = config.NewSelector(o.Text)
	if err != nil {
		return false, err
	}
	o.font, err = config.NewSelector(o.Font)
	if err != nil {
		return false, err
	}
	o.width, err = config.NewSelector(o.Width)
	if err != nil {
		return false, err
	}
	o.height, err = config.NewSelector(o.Height)
	if err != nil {
		return false, err
	}
	o.relativeDim, err = config.NewSelector(o.RelativeDim)
	if err != nil {
		return false, err
	}
	o.offsetX, err = config.NewSelector(o.OffsetX)
	if err != nil {
		return false, err
	}
	o.offsetY, err = config.NewSelector(o.OffsetY)
	if err != nil {
		return false, err
	}
	o.relativeOffset, err = config.NewSelector(o.RelativeOffset)
	if err != nil {
		return false, err
	}
	o.opacity, err = config.NewSelector(o.Opacity)
	if err != nil {
		return false, err
	}
	o.colorR, err = config.NewSelector(o.Color.R)
	if err != nil {
		return false, err
	}
	o.colorG, err = config.NewSelector(o.Color.G)
	if err != nil {
		return false, err
	}
	o.colorB, err = config.NewSelector(o.Color.B)
	if err != nil {
		return false, err
	}
	o.alignment, err = config.NewSelector(o.Alignment)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (o *label) Apply(p *vips.TransformParams, data payload.Data) error {
	var err error
	label := vips.LabelParams{
		Text:      "",
		Font:      "",
		Width:     vips.Scalar{},
		Height:    vips.Scalar{},
		OffsetX:   vips.Scalar{},
		OffsetY:   vips.Scalar{},
		Opacity:   0,
		Color:     vips.Color{},
		Alignment: 0,
	}

	//	//	//	//	//	//	//	//	//	//	//	//	//	//

	label.Text, err = o.text.Evaluate(data)
	if err != nil {
		return nil
	}

	label.Font, err = o.font.Evaluate(data)
	if err != nil {
		return nil
	}

	//	//	//	//	//	//	//	//	//	//	//	//	//	//

	w, err := o.width.EvaluateFloat64(data)
	if err != nil {
		return nil
	}

	h, err := o.height.EvaluateFloat64(data)
	if err != nil {
		return nil
	}

	relative, err := o.relativeDim.EvaluateBool(data)
	if err != nil {
		return nil
	}

	label.Width = vips.Scalar{
		Value:    w,
		Relative: relative,
	}
	label.Height = vips.Scalar{
		Value:    h,
		Relative: relative,
	}

	//	//	//	//	//	//	//	//	//	//	//	//	//	//

	x, err := o.offsetX.EvaluateFloat64(data)
	if err != nil {
		return nil
	}

	y, err := o.offsetY.EvaluateFloat64(data)
	if err != nil {
		return nil
	}

	relativeO, err := o.relativeOffset.EvaluateBool(data)
	if err != nil {
		return nil
	}

	label.OffsetX = vips.Scalar{
		Value:    x,
		Relative: relativeO,
	}
	label.OffsetY = vips.Scalar{
		Value:    y,
		Relative: relativeO,
	}

	//	//	//	//	//	//	//	//	//	//	//	//	//	//	//

	f64Opacity, err := o.opacity.EvaluateFloat64(data)
	if err != nil {
		return nil
	}

	label.Opacity = float32(f64Opacity)

	//	//	//	//	//	//	//	//	//	//	//	//	//	//	//

	label.Color.R, err = o.colorR.EvaluateUint8(data)
	if err != nil {
		return nil
	}

	label.Color.G, err = o.colorG.EvaluateUint8(data)
	if err != nil {
		return nil
	}

	label.Color.B, err = o.colorB.EvaluateUint8(data)
	if err != nil {
		return nil
	}

	//	//	//	//	//	//	//	//	//	//	//	//	//	//	//

	align, err := o.alignment.Evaluate(data)
	if err != nil {
		return nil
	}

	switch align {
	case "low":
		label.Alignment = vips.AlignLow
	case "center":
		label.Alignment = vips.AlignCenter
	case "high":
		label.Alignment = vips.AlignHigh
	default:
		return fmt.Errorf("invalid value for field [align], got: %s", align)
	}

	//	//	//	//	//	//	//	//	//	//	//	//	//	//	//

	// Apply some defaults
	if label.Width.IsZero() {
		label.Width.SetScale(1)
	}
	if label.Height.IsZero() {
		label.Height.SetScale(1)
	}
	if label.Font == "" {
		label.Font = vips.DefaultFont
	}
	if label.Opacity == 0 {
		label.Opacity = 1
	}

	// Set Transform params label to the label we just constructed
	p.Label = &label

	return err
}
