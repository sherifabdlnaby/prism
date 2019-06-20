package detectors

import (
	"bytes"
	"fmt"
	pigo "github.com/NohaSayedA/pigo/core"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"log"

	"github.com/sherifabdlnaby/prism/pkg/response"

	cfg "github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"

	"go.uber.org/zap"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

//FaceDetector plugin will read an image and identify human faces in image,
// and will draw a box around each person that was identified.
type FaceDetector struct {
	logger zap.SugaredLogger
	config config
}

//NewFaceDetector Return a new Component
func NewFaceDetector() component.Component {
	return &FaceDetector{}
}

// Init file validator
func (d *FaceDetector) Init(config cfg.Config, logger zap.SugaredLogger) error {
	d.config = *defaultConfig()
	var err error
	err = config.Populate(&d.config)
	if err != nil {
		return err
	}
	d.logger = logger
	return nil
}

// Start validator plugin
func (d *FaceDetector) Start() error {
	return nil
}

// Close validator plugin
func (d *FaceDetector) Close() error {
	return nil
}

// Decode return a decoded header(config) and image format from input bytes
func (d *FaceDetector) Decode(in payload.Bytes, data payload.Data) (payload.DecodedImage, response.Response) {
	return d.DecodeStream(bytes.NewReader(in), data)
}

// DecodeStream return a decoded header(config) and image format from input stream
func (d *FaceDetector) DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response) {
	reader := in.(io.Reader)
	image, _, err := image.Decode(reader)
	if err != nil {
		return nil, response.Error(fmt.Errorf("unsupported format"))
	}

	return image, response.Ack()
}

// Process Process will process the image and recognize the face in image, according to configuration.
func (d *FaceDetector) Process(in payload.DecodedImage, data payload.Data) (payload.DecodedImage, response.Response) {
	image := in.(image.Image)
	fd := pigo.NewFaceDetector(d.config.cascadeFile, d.config.minSize, d.config.maxSize, d.config.shiftFactor, d.config.scaleFactor, d.config.iouThreshold, d.config.angle)
	faces, err := fd.DetectFaces(image)
	if err != nil {
		log.Fatalf("Detection error: %v", err)
	}
	image = fd.DrawFaces(faces, d.config.circleMarker)

	return image, response.Ack()
}

// Encode will encode Image according to configuration, only supporting encoding into jpeg/png
func (d *FaceDetector) Encode(in payload.DecodedImage, data payload.Data) (payload.Bytes, response.Response) {
	image := in.(image.Image)
	outBuffer := bytes.Buffer{}
	var err error

	switch d.config.Export.Format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&outBuffer, image, &jpeg.Options{
			Quality: d.config.Export.Quality,
		})
	case "png":
		err = png.Encode(&outBuffer, image)
	}

	if err != nil {
		return nil, response.Error(err)
	}

	return outBuffer.Bytes(), response.Ack()
}
