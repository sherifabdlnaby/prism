package vips

import (
	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type blur struct {
	Sigma string
	sigma config.Selector
}

func (o *blur) IsActive() bool {
	return o.sigma.IsDynamic() || o.Sigma != ""
}

func (o *blur) Init() error {
	var err error

	o.sigma, err = config.NewSelector(o.Sigma)
	if err != nil {
		return err
	}

	return nil
}

func (o *blur) Apply(p *vips.TransformParams, data payload.Data) error {

	sigma, err := o.sigma.EvaluateFloat64(data)

	if err != nil {
		return err
	}

	p.BlurSigma = sigma

	return err
}
