package nude

import (
	"bytes"
	"fmt"
	"image"
	// register GIF to decode function
	_ "image/gif"
	// register JPEG to decode function
	_ "image/jpeg"
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

// Detector plugin will read an image and comes-up with the likely hood that this image contains nudity,
// Detector is a read-only plugin with no output. it can be configured to send a NoAck when it detects nudity or just
// add a flag to payload.Data
type Detector struct {
	logger zap.SugaredLogger
	config config
}

// NewDetector Return a new Detector Base
func NewDetector() component.Base {
	return &Detector{}
}

// Init file validator
func (d *Detector) Init(config cfg.Config, logger zap.SugaredLogger) error {
	var err error

	d.config = *defaultConfig()
	err = config.Populate(&d.config)
	if err != nil {
		return err
	}

	d.logger = logger
	return nil
}

// Start validator plugin
func (d *Detector) Start() error {
	return nil
}

// Stop validator plugin
func (d *Detector) Stop() error {
	return nil
}

// Decode return a decoded header(config) and image format from input bytes
func (d *Detector) Decode(in payload.Bytes, data payload.Data) (payload.DecodedImage, response.Response) {
	return d.DecodeStream(bytes.NewReader(in), data)
}

// DecodeStream return a decoded header(config) and image format from input stream
func (d *Detector) DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response) {
	reader := in.(io.Reader)

	image, _, err := image.Decode(reader)
	if err != nil {
		return nil, response.Error(fmt.Errorf("unsupported format"))
	}

	return image, response.Ack()
}

// process process will process the image and calculate skin regions and the likelihood it's a nude image, according to
// configuration, process will either send a NoAck to Drop the image according to configuration,
// otherwise will add "nude" boolean to payload.Data
func (d *Detector) Process(in payload.DecodedImage, data payload.Data) response.Response {

	image := in.(image.Image)
	decoder := gonude.NewDetector(image)

	isNude, err := decoder.Parse()
	if err != nil {
		return response.Error(err)
	}

	// add to data
	data["nude_detection"] = isNude

	if d.config.Drop && isNude {
		return response.NoAck(fmt.Errorf("image may contain nudity"))
	}

	return response.Ack()

}
