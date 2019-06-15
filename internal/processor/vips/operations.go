package vips

import (
	"github.com/h2non/bimg"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type Operations struct {
	// Parsing
	Resize resize
	Flip   flip
	Blur   blur
	Rotate rotate
	Crop   crop
	Label  label
	//------------disabled-------------------
	//Invert  invert	`mapstructure:",squash"`
	//---------------------------------------
	// for internal use
	operations []operation
}

type operation interface {
	Init() (bool, error)
	Apply(p *bimg.Options, data payload.Data) error
}

func (o *Operations) Init() error {

	// Init every operation and add them if they're active.
	ok, err := o.Resize.Init()
	if err != nil {
		return err
	}
	if ok {
		o.operations = append(o.operations, &o.Resize)
	}

	// Init every operation and add them if they're active.
	ok, err = o.Flip.Init()
	if err != nil {
		return err
	}
	if ok {
		o.operations = append(o.operations, &o.Flip)
	}

	// Init every operation and add them if they're active.
	ok, err = o.Blur.Init()
	if err != nil {
		return err
	}
	if ok {
		o.operations = append(o.operations, &o.Blur)
	}

	// Init every operation and add them if they're active.
	ok, err = o.Rotate.Init()
	if err != nil {
		return err
	}
	if ok {
		o.operations = append(o.operations, &o.Rotate)
	}

	//// Init every operation and add them if they're active.
	//ok, err = o.Scale.Init()
	//if err != nil {
	//	return err
	//}
	//if ok {
	//	o.operations = append(o.operations, &o.Scale)
	//}
	//
	// Init every operation and add them if they're active.
	ok, err = o.Crop.Init()
	if err != nil {
		return err
	}
	if ok {
		o.operations = append(o.operations, &o.Crop)
	}

	// Init every operation and add them if they're active.
	ok, err = o.Label.Init()
	if err != nil {
		return err
	}
	if ok {
		o.operations = append(o.operations, &o.Label)
	}

	// Init every operation and add them if they're active.
	//	ok, err = o.Invert.Init()
	//	if err != nil {
	//		return err
	//	}
	// if ok {
	// 	o.operations = append(o.operations, &o.Resize)
	// }

	return nil
}

func (o *Operations) Do(params *bimg.Options, data payload.Data) error {
	var err error

	// Apply operations to params
	for _, op := range o.operations {
		err = op.Apply(params, data)
		if err != nil {
			return err
		}
	}

	return nil
}
