package vips

import (
	"github.com/sherifabdlnaby/bimg"
	cfg "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

// TODO : ONLY PNG
type label struct {
	Raw       labelRawConfig `mapstructure:",squash"`
	width     cfg.Selector
	dpi       cfg.Selector
	margin    cfg.Selector
	opacity   cfg.Selector
	replicate cfg.Selector
	text      cfg.Selector
	font      cfg.Selector
	colorR    cfg.Selector
	colorG    cfg.Selector
	colorB    cfg.Selector
}

type labelRawConfig struct {
	Width     string
	DPI       string
	Margin    string
	Opacity   string
	Replicate string
	Text      string
	Font      string
	Color     rgb
}

type rgb struct {
	R, G, B string
}

func (o *label) Init() (bool, error) {
	var err error

	if o.Raw == *labelDefaults() {
		return false, nil
	}

	o.text, err = cfg.NewSelector(o.Raw.Text)
	if err != nil {
		return false, err
	}
	o.font, err = cfg.NewSelector(o.Raw.Font)
	if err != nil {
		return false, err
	}
	o.width, err = cfg.NewSelector(o.Raw.Width)
	if err != nil {
		return false, err
	}
	o.dpi, err = cfg.NewSelector(o.Raw.DPI)
	if err != nil {
		return false, err
	}
	o.opacity, err = cfg.NewSelector(o.Raw.Opacity)
	if err != nil {
		return false, err
	}
	o.margin, err = cfg.NewSelector(o.Raw.Margin)
	if err != nil {
		return false, err
	}
	o.replicate, err = cfg.NewSelector(o.Raw.Replicate)
	if err != nil {
		return false, err
	}

	o.colorR, err = cfg.NewSelector(o.Raw.Color.R)
	if err != nil {
		return false, err
	}
	o.colorG, err = cfg.NewSelector(o.Raw.Color.G)
	if err != nil {
		return false, err
	}
	o.colorB, err = cfg.NewSelector(o.Raw.Color.B)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (o *label) Apply(p *bimg.Options, data payload.Data) error {
	var err error
	label := bimg.Watermark{
		Width:       0,
		DPI:         0,
		Margin:      0,
		Opacity:     0,
		NoReplicate: false,
		Text:        "",
		Font:        "",
		Background:  bimg.Color{},
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
	label.Width = int(w)

	dpi, err := o.dpi.EvaluateInt64(data)
	if err != nil {
		return nil
	}
	label.DPI = int(dpi)

	margin, err := o.margin.EvaluateInt64(data)
	if err != nil {
		return nil
	}
	label.Margin = int(margin)

	op, err := o.opacity.EvaluateFloat64(data)
	if err != nil {
		return nil
	}
	label.Opacity = float32(op)

	replicate, err := o.replicate.EvaluateBool(data)
	if err != nil {
		return nil
	}
	label.NoReplicate = !replicate

	//	//	//	//	//	//	//	//	//	//	//	//	//	//	//

	label.Background.R, err = o.colorR.EvaluateUint8(data)
	if err != nil {
		return nil
	}

	label.Background.G, err = o.colorG.EvaluateUint8(data)
	if err != nil {
		return nil
	}

	label.Background.B, err = o.colorB.EvaluateUint8(data)
	if err != nil {
		return nil
	}

	p.Watermark = label

	return err
}
