package dummy

import (
	"errors"
	"fmt"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
	"golang.org/x/image/webp"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

//Dummy Dummy ProcessReadWrite that does absolutely nothing to the image
type Dummy struct {
	logger zap.SugaredLogger
}

type internalImage struct {
	Internal  []byte
	Name      string
	Image     *customReader
	ImageData payload.Data
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &Dummy{}
}

// Decode Simulate Decoding the Image
func (d *Dummy) Decode(in payload.Bytes, data payload.Data) (payload.DecodedImage, response.Response) {
	// create internal decoded object (varies with each plugin)`
	out := newCustomReaderBuffer(in)

	return internalImage{
		//Name:      ep.Name,
		Image:     out,
		ImageData: data,
	}, response.ACK
}

// DecodeStream Simulate Decoding the Image
func (d *Dummy) DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response) {
	out := newCustomReader(in)

	return internalImage{
		//Name:      ep.Name,
		Image:     out,
		ImageData: data,
	}, response.ACK
}

//Process Simulate Processing the Image
func (d *Dummy) Process(in payload.DecodedImage, data payload.Data) (payload.DecodedImage, response.Response) {
	//d.logger.Info("Processing Payload... ", zap.String("name", in.Name))

	inImage, ok := in.(internalImage)
	if !ok {
		return inImage, response.Error(errors.New("unable to cast to internalImage"))
	}

	// We will read from the tee reader
	// and save everything thing we read in buf
	//var buf bytes.Buffer
	//tee := io.TeeReader(inImage.Image, &buf)

	// Here we add the validation function for each type
	fn := []func(io.Reader) (image.Config, error){
		jpeg.DecodeConfig,
		png.DecodeConfig,
		webp.DecodeConfig,
	}

	for i, f := range fn {
		conf, err := f(inImage.Image)
		// Success
		if err == nil {
			// ReadAll to move all the file to buf
			//_, _ = ioutil.ReadAll(tee)

			d.logger.Debug("Image type found")
			d.logger.Debug(fmt.Sprintf("Image Info: %s %v*%v",
				imageType(i).String(), conf.Height, conf.Width))

			inImage.Image.Reset()
			inImage.ImageData = map[string]interface{}{
				"type":   imageType(i).String(),
				"height": conf.Height,
				"width":  conf.Width,
			}
			return inImage, response.ACK
		}
		inImage.Image.Reset()
	}

	//_, _ = ioutil.ReadAll(tee)
	d.logger.Error("Couldn't define image type")
	//return dp, errors.New("unidentified type")
	return inImage, response.ACK
}

// Encode Simulate Encoding the Image
func (d *Dummy) Encode(in payload.DecodedImage, data payload.Data) (payload.Bytes, response.Response) {
	// Since in this dummy case we have processed output as a whole, we can just pass it to next node.
	Payload := in.(internalImage).Internal

	return Payload, response.ACK
}

//Init Initialize Plugin based on parsed config
func (d *Dummy) Init(config config.Config, logger zap.SugaredLogger) error {
	d.logger = logger
	return nil
}

//Start start the plugin to begin receiving input
func (d *Dummy) Start() error {
	d.logger.Info("Started Dummy processor.")
	return nil
}

//Close Close plugin gracefully
func (d *Dummy) Close() error {
	return nil
}
