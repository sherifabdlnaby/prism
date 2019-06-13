package vips

import (
	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

// TODO : This is NOT Used because it outputs black images on most pictures
type invert struct {
	Invert string
	invert config.Selector
}

func (o *invert) Init() (bool, error) {
	var err error

	//TODO check defaults

	o.invert, err = config.NewSelector(o.Invert)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (o *invert) Apply(p *vips.TransformParams, data payload.Data) error {

	invert, err := o.invert.EvaluateBool(data)

	if err != nil {
		return err
	}

	p.Invert = invert

	return err
}
