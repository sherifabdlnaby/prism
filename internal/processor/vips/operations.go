package vips

import (
	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type Operations struct {
	// Parsing
	Resize resize
	Flip   flip

	// for internal use
	operations []operation
}

type operation interface {
	Init() error
	IsActive() bool
	Apply(p *vips.TransformParams, data payload.Data) error
}

func (o *Operations) Init() error {

	// Init every operation and add them if they're active.
	if o.Resize.IsActive() {
		err := o.Resize.Init()
		if err != nil {
			return err
		}
		o.operations = append(o.operations, &o.Resize)
	}

	// Init every operation and add them if they're active.
	if o.Flip.IsActive() {
		err := o.Flip.Init()
		if err != nil {
			return err
		}
		o.operations = append(o.operations, &o.Flip)
	}

	return nil
}

func (o *Operations) Do(image *vips.ImageRef, data payload.Data) error {
	var err error

	// build transform params
	params := DefaultTransformParams()

	// Apply operations to params
	for _, op := range o.operations {
		err = op.Apply(params, data)
		if err != nil {
			return err
		}
	}

	// Create Backboard ( govips internal structure that eases transformations
	bb := vips.NewBlackboard(image.Image(), image.Format(), params)
	err = vips.ProcessBlackboard(bb)
	image.SetImage(bb.Image())

	if err != nil {
		return err
	}

	return nil
}
