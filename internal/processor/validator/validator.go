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

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
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

	// check if we'll only need to get format only (to use a faster method)
	if d.config.MinHeight+d.config.MaxHeight+d.config.MinWidth+d.config.MaxWidth == 0 {
		d.config.formatOnly = true
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

// Decode Decodes an image reading only the necessary bytes to validate the image
func (d *Validator) Decode(in payload.Bytes, data payload.Data) (payload.DecodedImage, response.Response) {
	return d.DecodeStream(bytes.NewReader(in), data)
}

// DecodeStream Decodes an image reading only the necessary bytes to validate the image
func (d *Validator) DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response) {
	reader := in.(io.Reader)

	var format string

	// if only need to check for type -> use the quicker method that only need 260 byte
	if d.config.formatOnly {
		head := make([]byte, 261)
		_, err := io.ReadFull(reader, head)
		if err != nil {
			return nil, response.NoAck(fmt.Errorf("bytes not enough to validate image type"))
		}

		if jpeg := filetype.IsType(head, matchers.TypeJpeg); jpeg {
			format = "jpeg"
		} else if png := filetype.IsType(head, matchers.TypePng); png {
			format = "png"
		} else if webp := filetype.IsType(head, matchers.TypeWebp); webp {
			format = "webp"
		}

		return header{
			format: format,
		}, response.Ack()
	}

	// use image.Decode to get both format AND dimensions
	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, response.NoAck(fmt.Errorf("unsupported format"))
	}

	return header{
		format: format,
		Config: config,
	}, response.Ack()
}

// Process will validate that the image is as configured. adding format and dimensions to payload.Dataa
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

	data["_format"] = header.format

	if !d.config.formatOnly {
		data["_width"] = header.Width
		data["_height"] = header.Height
	}

	if header.Width > d.config.MaxWidth ||
		header.Width < d.config.MinWidth ||
		header.Height > d.config.MaxHeight ||
		header.Height < d.config.MinHeight {
		return response.NoAck(fmt.Errorf("unsupported format"))
	}

	return response.Ack()
}
