package validator

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

	"github.com/sherifabdlnaby/prism/pkg/component"

	// register webp to decode function
	_ "golang.org/x/image/webp"

	cfg "github.com/sherifabdlnaby/prism/pkg/config"

	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

// Validator plugin reads the first n bytes necessary to get image-format and its header and use them for validation.
type Validator struct {
	logger zap.SugaredLogger
	config config
}

type header struct {
	format string
	image.Config
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &Validator{}
}

// Init file validator
func (d *Validator) Init(config cfg.Config, logger zap.SugaredLogger) error {
	var err error

	d.config = *defaultConfig()
	err = config.Populate(&d.config)
	if err != nil {
		return err
	}

	// set enabled types
	for _, value := range d.config.Format {
		switch value {
		case "jpeg", "jpg":
			d.config.jpeg = true
		case "png":
			d.config.png = true
		case "webp":
			d.config.webp = true
		default:
			return fmt.Errorf("unsuppprted format configuration [%s]", value)
		}
	}

	d.logger = logger
	return nil
}

// Start validator plugin
func (d *Validator) Start() error {
	return nil
}

// Close validator plugin
func (d *Validator) Close() error {
	return nil
}

// Decode return a decoded header(config) and image format from input bytes
func (d *Validator) Decode(in payload.Bytes, data payload.Data) (payload.DecodedImage, response.Response) {
	return d.DecodeStream(bytes.NewReader(in), data)
}

// DecodeStream return a decoded header(config) and image format from input stream
func (d *Validator) DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response) {
	reader := in.(io.Reader)

	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, response.NoAck(fmt.Errorf("unsupported format"))
	}

	return header{
		format: format,
		Config: config,
	}, response.Ack()
}

//Process Compare decoded header and format with configuration
func (d *Validator) Process(in payload.DecodedImage, data payload.Data) response.Response {
	var err error
	header := in.(header)

	switch header.format {
	case "jpeg":
		if !d.config.jpeg {
			err = fmt.Errorf("unsupported format")
		}
	case "png":
		if !d.config.png {
			err = fmt.Errorf("unsupported format")
		}
	case "webp":
		if !d.config.webp {
			err = fmt.Errorf("unsupported format")
		}
	default:
		err = fmt.Errorf("unsupported format")
	}

	if err != nil {
		return response.NoAck(err)
	}

	if header.Width > d.config.MaxWidth ||
		header.Width < d.config.MinWidth ||
		header.Height > d.config.MaxHeight ||
		header.Height < d.config.MinHeight {
		return response.NoAck(fmt.Errorf("unsupported image dimenstions"))
	}

	return response.Ack()

}
