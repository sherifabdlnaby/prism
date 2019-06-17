package vips

import (
	"github.com/sherifabdlnaby/bimg"
	cfg "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type blur struct {
	Raw    blurRawConfig `mapstructure:",squash"`
	sigma  cfg.Selector
	minAmp cfg.Selector
}

type blurRawConfig struct {
	Sigma   string
	MinAmpl string `mapstructure:"min_ampl"`
}

func (o *blur) Init() (bool, error) {
	var err error

	if o.Raw == *blurDefaults() {
		return false, nil
	}

	o.sigma, err = cfg.NewSelector(o.Raw.Sigma)
	if err != nil {
		return false, err
	}

	o.minAmp, err = cfg.NewSelector(o.Raw.MinAmpl)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (o *blur) Apply(p *bimg.Options, data payload.Data) error {

	sigma, err := o.sigma.EvaluateFloat64(data)

	if err != nil {
		return err
	}

	minAmp, err := o.minAmp.EvaluateFloat64(data)

	if err != nil {
		return err
	}

	p.GaussianBlur = bimg.GaussianBlur{
		Sigma:   sigma,
		MinAmpl: minAmp,
	}

	return err
}
