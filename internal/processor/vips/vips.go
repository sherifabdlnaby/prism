package vips

import (
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

//Dummy Dummy Processor that does absolutely nothing to the image
type Vips struct {
	logger zap.SugaredLogger
	config Config
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &Vips{}
}

type internalImage struct {
	internal *vips.ImageRef
}

type Config struct {
	Operations Operations
}

func DefaultConfig() *Config {
	return &Config{
		Operations{
			Resize: resize{
				Strategy: "auto",
				Pad:      "black",
			},
			Flip: flip{
				Direction: "none",
			},
		},
	}
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
	defer runtime.KeepAlive(in)

	img, err := vips.NewImageFromBuffer(in)

	if err != nil {
		return nil, response.Error(err)
	}

	if img.Format() == vips.ImageTypeUnknown {
		return nil, response.Error(fmt.Errorf("unknown image type"))
	}

	// create internal object (varies with each plugin)`
	out := internalImage{
		internal: img,
	}

	//TODO decode according to type

	// Return it as it is (dummy).
	return out, response.ACK
}

func (d *Vips) DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response) {
	buff, err := ioutil.ReadAll(in)

	if err != nil {
		return nil, response.Error(err)
	}

	defer runtime.KeepAlive(buff)

	img, err := vips.NewImageFromBuffer(buff)

	if err != nil {
		return nil, response.Error(err)
	}

	if img.Format() == vips.ImageTypeUnknown {
		return nil, response.Error(fmt.Errorf("unknown image type"))
	}

	// create internal object (varies with each plugin)`
	out := internalImage{
		internal: img,
	}

	// Return it as it is (dummy).
	return out, response.ACK
}

func (d *Vips) Process(in payload.DecodedImage, data payload.Data) (payload.DecodedImage, response.Response) {
	origImg := in.(internalImage)
	img := vips.NewImageFromRef(origImg.internal)

	err := d.config.Operations.Do(img, data)

	if err != nil {
		return nil, response.Error(err)
	}

	return internalImage{
		internal: img,
	}, response.ACK
}

func (d *Vips) Encode(in payload.DecodedImage, data payload.Data) (payload.Bytes, response.Response) {
	Img := in.(internalImage)

	byteBuff, _, err := Img.internal.Export(vips.ExportParams{
		Format:  vips.ImageTypeJPEG,
		Quality: 95,
	})

	if err != nil {
		return nil, response.Error(err)
	}

	return byteBuff, response.ACK
}

func init() {
	vips.Startup(&vips.Config{
		ConcurrencyLevel: 4,
	})
}
