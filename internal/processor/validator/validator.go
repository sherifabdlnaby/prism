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

//Validator struct
//defines file type
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

//Init file validator
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

//Start the validator plugin
func (d *Validator) Start() error {
	return nil
}

//Close the validator plugin
func (d *Validator) Close() error {
	return nil
}

//Decode
func (d *Validator) Decode(in payload.Bytes, data payload.Data) (payload.DecodedImage, response.Response) {
	return d.DecodeStream(bytes.NewReader(in), data)
}

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

func (d *Validator) Process(in payload.DecodedImage, data payload.Data) response.Response {
	header := in.(header)

	switch header.format {
	case "jpeg":
		if !d.config.jpeg {
			return response.NoAck(fmt.Errorf("unsupported format"))
		}
	case "png":
		if !d.config.png {
			return response.NoAck(fmt.Errorf("unsupported format"))
		}
	case "webp":
		if !d.config.webp {
			return response.NoAck(fmt.Errorf("unsupported format"))
		}
	default:
		return response.NoAck(fmt.Errorf("unsupported format"))
	}

	if header.Width > d.config.MaxWidth ||
		header.Width < d.config.MinWidth ||
		header.Height > d.config.MaxHeight ||
		header.Height < d.config.MinHeight {
		return response.NoAck(fmt.Errorf("unsupported format"))
	}

	// Add format to Data
	data["format"] = header.format

	return response.Ack()

}
