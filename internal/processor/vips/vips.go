package vips

import (
	"io/ioutil"
	"runtime"

	"github.com/sherifabdlnaby/bimg"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

// TODO: EXTEND BIMG TO DO FACE AND ENTROPY smart crop

func init() {
	//bimg.VipsCacheSetMaxMem(0)
	//bimg.VipsCacheSetMax(0)
}

//Dummy Dummy Processor that does absolutely nothing to the image
type Vips struct {
	logger zap.SugaredLogger
	config Config
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &Vips{}
}

type image struct {
	image   *bimg.VipsImage
	options bimg.Options
}

//Init Initialize Plugin based on parsed Operations
func (d *Vips) Init(config config.Config, logger zap.SugaredLogger) error {
	var err error

	d.config = *DefaultConfig()
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
	d.logger.Info("Started Dummy processor.")
	return nil
}

//Close Close plugin gracefully
func (d *Vips) Close() error {
	return nil
}

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

func (d *Vips) DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response) {
	buff, err := ioutil.ReadAll(in)

	vimage, err := bimg.NewVipsImage(buff)

	if err != nil {
		return nil, response.Error(err)
	}

	return image{
		image:   vimage,
		options: bimg.Options{},
	}, response.ACK
}

func (d *Vips) Process(in payload.DecodedImage, data payload.Data) (payload.DecodedImage, response.Response) {
	// TODO fix bimg fork to not need this
	//  (this happen probably because shrink on load requires src buffer to still be alive )
	defer runtime.KeepAlive(in)

	vimage := in.(image)
	params := defaultOptions()

	img := vimage.image.Clone()

	// apply configs
	err := d.config.Operations.Do(params, data)
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

	return bytes, response.ACK
}
