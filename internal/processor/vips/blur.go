package vips

import (
	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type blur struct {
	Raw   blurRawConfig `mapstructure:",squash"`
	sigma config.Selector
}

type blurRawConfig struct {
	Sigma string
}

func (o *blur) Init() (bool, error) {
	var err error

	if o.Raw == *blurDefaults() {
		return false, nil
	}

	o.sigma, err = config.NewSelector(o.Raw.Sigma)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (o *blur) Apply(p *vips.TransformParams, data payload.Data) error {

	sigma, err := o.sigma.EvaluateFloat64(data)

	if err != nil {
		return err
	}

	p.BlurSigma = sigma

	return err
}
