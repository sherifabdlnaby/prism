package nude

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	// register GIF to decode function
	_ "image/gif"
	"image/jpeg"
	// register JPEG to decode function
	_ "image/jpeg"
	"image/png"
	// register PNG to decode function
	_ "image/png"
	"io"

	"github.com/sherifabdlnaby/prism/internal/processor/nude/gonude"
	"github.com/sherifabdlnaby/prism/pkg/component"

	// register webp to decode function
	_ "golang.org/x/image/webp"

	cfg "github.com/sherifabdlnaby/prism/pkg/config"

	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

// Censor plugin will read an image and comes-up with the likely hood that this image contains nudity,
// and will censor the image with configured color pixels. It can be configured to send a NoAck when it detects nudity and
// will a boolean flag "nude" to payload.Data
type Censor struct {
	logger zap.SugaredLogger
	config config
}

// NewCensor Return a new Censor Base
func NewCensor() component.Base {
	return &Censor{}
}

// Init file validator
func (d *Censor) Init(config cfg.Config, logger zap.SugaredLogger) error {
	var err error

	d.config = *defaultConfig()
	err = config.Populate(&d.config)
	if err != nil {
		return err
	}

	d.config.rgba = color.RGBA{
		R: d.config.RGBA.R,
		G: d.config.RGBA.G,
		B: d.config.RGBA.B,
		A: uint8(d.config.RGBA.A * 255),
	}

	d.logger = logger
	return nil
}

// Start validator plugin
func (d *Censor) Start() error {
	return nil
}

// Stop validator plugin
func (d *Censor) Stop() error {
	return nil
}

// Decode return a decoded header(config) and image format from input bytes
func (d *Censor) Decode(in payload.Bytes, data payload.Data) (payload.DecodedImage, response.Response) {
	return d.DecodeStream(bytes.NewReader(in), data)
}

// DecodeStream return a decoded header(config) and image format from input stream
func (d *Censor) DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response) {
	reader := in.(io.Reader)

	image, _, err := image.Decode(reader)
	if err != nil {
		return nil, response.Error(fmt.Errorf("unsupported format"))
	}

	return image, response.Ack()
}

// Process Process will process the image and calculate skin regions and the likelihood it's a nude image, according to
// configuration, process will either send a NoAck to Drop the image, OR censor the image by adding pixels over skin regions
// will also add "nude" boolean to payload.Data
func (d *Censor) Process(in payload.DecodedImage, data payload.Data) (payload.DecodedImage, response.Response) {
	image := in.(image.Image)

	decoder := gonude.NewDetector(image)

	isNude, err := decoder.Parse()
	if err != nil {
		return nil, response.Error(err)
	}

	// add to data
	data["nude_detection"] = isNude

	if d.config.Drop && isNude {
		return nil, response.NoAck(fmt.Errorf("image may contain nudity"))
	}

	if isNude {
		image = gonude.CensorSkinPixels(image, decoder.SkinRegions, d.config.rgba)
	}

	return image, response.Ack()
}

// Encode will encode Image according to configuration, only supporting encoding into jpeg/png
func (d *Censor) Encode(in payload.DecodedImage, data payload.Data) (payload.Bytes, response.Response) {
	var err error
	img := in.(image.Image)

	outBuffer := bytes.Buffer{}

	switch d.config.Export.Format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&outBuffer, img, &jpeg.Options{
			Quality: d.config.Export.Quality,
		})
	case "png":
		err = png.Encode(&outBuffer, img)
	}

	if err != nil {
		return nil, response.Error(err)
	}

	// add to data
	max := img.Bounds().Max

	data["_format"] = d.config.Export.Format
	data["_width"] = max.X
	data["_height"] = max.Y

	return outBuffer.Bytes(), response.Ack()
}
