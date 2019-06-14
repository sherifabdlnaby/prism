package vips

import (
	"io/ioutil"

	"github.com/h2non/bimg"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

// TODO: Memory leak occurs when data from transactions are bytes not stream from INPUT, AND there are tons of requests (10ms between each request)
// TODO: EXTEND BIMG TO DO FACE AND ENTROPY smart crop

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
	bytes   []byte
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
	d.config.export, err = NewExportParams(d.config.Export)
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
	bimg.VipsCacheDropAll()
	return nil
}

func (d *Vips) Decode(in payload.Bytes, data payload.Data) (payload.DecodedImage, response.Response) {
	// Return it as it is (dummy).
	return []byte(in), response.ACK
}

func (d *Vips) DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response) {
	buff, err := ioutil.ReadAll(in)

	if err != nil {
		return nil, response.Error(err)
	}

	// Return it as it is (dummy).
	return buff, response.ACK
}

func (d *Vips) Process(in payload.DecodedImage, data payload.Data) (payload.DecodedImage, response.Response) {
	params := defaultOptions()

	err := d.config.Operations.Do(params, data)

	if err != nil {
		return nil, response.Error(err)
	}

	return &image{
		bytes:   in.([]byte),
		options: *params,
	}, response.ACK
}

func (d *Vips) Encode(in payload.DecodedImage, data payload.Data) (payload.Bytes, response.Response) {
	img := in.(*image)
	bytes, err := bimg.Resize(img.bytes, img.options)

	if err != nil {
		return nil, response.Error(err)
	}

	return bytes, response.ACK
}
