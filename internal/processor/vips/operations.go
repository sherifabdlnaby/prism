package vips

import (
	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type Operations struct {
	// Parsing
	Resize resize
	Flip   flip
	Blur   blur
	Rotate rotate
	Scale  scale
	//------------disabled-------------------
	//Label  label
	//Invert  invert	`mapstructure:",squash"`
	//---------------------------------------
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

	// Init every operation and add them if they're active.
	if o.Blur.IsActive() {
		err := o.Blur.Init()
		if err != nil {
			return err
		}
		o.operations = append(o.operations, &o.Blur)
	}

	// Init every operation and add them if they're active.
	if o.Rotate.IsActive() {
		err := o.Rotate.Init()
		if err != nil {
			return err
		}
		o.operations = append(o.operations, &o.Rotate)
	}

	// Init every operation and add them if they're active.
	if o.Scale.IsActive() {
		err := o.Scale.Init()
		if err != nil {
			return err
		}
		o.operations = append(o.operations, &o.Scale)
	}

	//// Init every operation and add them if they're active.
	//if o.Label.IsActive() {
	//	err := o.Label.Init()
	//	if err != nil {
	//		return err
	//	}
	//	o.operations = append(o.operations, &o.Label)
	//}

	// Init every operation and add them if they're active.
	//if o.Invert.IsActive() {
	//	err := o.Invert.Init()
	//	if err != nil {
	//		return err
	//	}
	//	o.operations = append(o.operations, &o.Invert)
	//}

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
