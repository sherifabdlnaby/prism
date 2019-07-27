package vips

import (
	"io/ioutil"
	"runtime"

	"github.com/sherifabdlnaby/bimg"
	"github.com/sherifabdlnaby/prism/pkg/component"
	cfg "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

// TODO: EXTEND BIMG TO DO FACE AND ENTROPY smart crop

// Vips a processing plugin that use libvips C library to do multiple operations
type Vips struct {
	logger zap.SugaredLogger
	config config
}

// NewComponent Return a new Base
func NewComponent() component.Base {
	return &Vips{}
}

type image struct {
	image   *bimg.VipsImage
	options bimg.Options
}

//Init Initialize Plugin based on parsed operations
func (d *Vips) Init(config cfg.Config, logger zap.SugaredLogger) error {
	var err error

	d.config = *defaultConfig()
	err = config.Populate(&d.config)
	if err != nil {
		return err
	}

	// init operations
	err = d.config.Operations.Init()
	if err != nil {
		return err
	}

	// init export
	_, err = d.config.Export.Init()
	if err != nil {
		return err
	}

	d.logger = logger
	return nil
}

//Start Start the plugin to begin receiving input
func (d *Vips) Start() error {
	return nil
}

//Stop Stop plugin gracefully
func (d *Vips) Stop() error {
	return nil
}

//Decode Decode input Bytes into a vipsImage
func (d *Vips) Decode(in payload.Bytes, data payload.Data) (payload.DecodedImage, response.Response) {

	vimage, err := bimg.NewVipsImage(in)

	if err != nil {
		return nil, response.Error(err)
	}

	return image{
		image:   vimage,
		options: bimg.Options{},
	}, response.ACK
}

//DecodeStream Decode input Stream into a vipsImage
func (d *Vips) DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response) {
	buff, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, response.Error(err)
	}

	vimage, err := bimg.NewVipsImage(buff)
	if err != nil {
		return nil, response.Error(err)
	}

	return image{
		image:   vimage,
		options: bimg.Options{},
	}, response.ACK
}

//Process Process vips image according to internal configuration of Vips plugin
func (d *Vips) Process(in payload.DecodedImage, data payload.Data) (payload.DecodedImage, response.Response) {
	// TODO fix bimg fork to not need this
	//  (this happen probably because shrink on load requires src buffer to still be alive )
	defer runtime.KeepAlive(in)

	vimage := in.(image)
	params := defaultOptions()

	img := vimage.image.Clone()

	//TODO use clone instead when finishing forking BIMG to be ready to use intermediate results.
	//img := vimage.image.Clone()

	// apply configs
	err := d.config.Operations.Apply(params, data)
	if err != nil {
		return nil, response.Error(err)
	}

	// process
	err = img.Process(*params)
	if err != nil {
		return nil, response.Error(err)
	}

	return &image{
		image:   img,
		options: *params,
	}, response.ACK
}

//Encode Encodes the image according to internal configurations of the plugin and returns it as a byte buffer
func (d *Vips) Encode(in payload.DecodedImage, data payload.Data) (payload.Bytes, response.Response) {
	defer runtime.KeepAlive(in)

	img := in.(*image)

	// apply export
	err := d.config.Export.Apply(&img.options, data)
	if err != nil {
		return nil, response.Error(err)
	}

	bytes, err := img.image.Save(img.options)
	if err != nil {
		return nil, response.Error(err)
	}

	data["_width"], data["_height"] = img.image.GetDimensions()
	data["_format"] = d.config.Export.Raw.Format

	return bytes, response.ACK
}
