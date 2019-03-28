package validator

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"golang.org/x/image/webp"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"time"
)

type Validator struct {
	logger zap.Logger
}

func (d *Validator) Decode(ep types.Payload) (types.DecodedPayload, error) {
	d.logger.Info("Decoding Payload... ", zap.String("name", ep.Name))

	// Return it as it is (dummy).
	return types.DecodedPayload{
		Name:      ep.Name,
		Image:     ep.Reader,
		ImageData: ep.ImageData,
	}, nil
}

func (d *Validator) Process(dp types.DecodedPayload) (types.DecodedPayload, error) {
	d.logger.Info("Processing Payload... ", zap.String("name", dp.Name))

	imageReader, ok := dp.Image.(io.Reader)
	if !ok {
		d.logger.Error("Invalid reader")
		return dp, errors.New("invalid reader")
	}

	//We will read from the tee reader
	//and save everything thing we read in buf
	var buf bytes.Buffer
	tee := io.TeeReader(imageReader, &buf)

	//Here we add the validation function for each type
	fn := []func(io.Reader) (image.Config, error){
		jpeg.DecodeConfig,
		png.DecodeConfig,
		webp.DecodeConfig,
	}

	myReader := NewReader(tee)

	for i, f := range fn {
		conf, err := f(myReader)
		if err == nil {
			//ReadAll to move all the file to buf
			_, _ = ioutil.ReadAll(tee)

			d.logger.Debug("Image type found")
			d.logger.Debug(fmt.Sprintf("Image Info: %s %v*%v",
				imageType(i).String(), conf.Height, conf.Width))

			dp.Image = &buf
			dp.ImageData = map[string]interface{}{
				"type": imageType(i).String(),
				"height": conf.Height,
				"width": conf.Width,
			}
			return dp, nil
		}
		myReader.Reset()
	}

	_, _ = ioutil.ReadAll(tee)
	dp.Image = &buf
	d.logger.Error("Couldn't define image type")
	return dp, errors.New("unidentified type")
}

func (d *Validator) Encode(dp types.DecodedPayload) (types.Payload, error) {
	d.logger.Info("Encoding Payload... ", zap.String("name", dp.Name))

	return types.Payload{
		Name:      dp.Name,
		Reader:    dp.Image.(io.Reader),
		ImageData: dp.ImageData,
	}, nil
}

func (d *Validator) Init(config types.Config, logger zap.Logger) error {
	d.logger = logger
	return nil
}

func (d *Validator) Start() error {
	d.logger.Info("Started Validator processor.")
	return nil
}

func (d *Validator) Close(time.Duration) error {
	return nil
}
