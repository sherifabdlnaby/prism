package dummy

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"go.uber.org/zap"
	"io/ioutil"
	"time"
)

type Dummy struct {
	logger zap.SugaredLogger
}

type internalImage struct {
	internal []byte
}

func NewComponent() component.Component {
	return &Dummy{}
}

func (d *Dummy) Decode(in component.InputPayload, data component.ImageData) (interface{}, component.Response) {
	d.logger.Debugw("Decoding InputPayload... ")

	imgBytes, err := ioutil.ReadAll(in)

	if err != nil {
		return component.DecodedPayload{}, component.ResponseError(err)
	}

	// create internal object (varies with each plugin)
	out := internalImage{
		internal: imgBytes,
	}

	// Return it as it is (dummy).
	return out, component.ResponseACK
}

func (d *Dummy) Process(dp interface{}, data component.ImageData) (interface{}, component.Response) {
	d.logger.Debugw("Processing InputPayload... ")
	return dp, component.ResponseACK
}

func (d *Dummy) Encode(in interface{}, data component.ImageData, out *component.OutputPayload) component.Response {
	d.logger.Debugw("Encoding InputPayload... ")
	out.ImageBytes = in.(internalImage).internal
	_, err := out.Write(out.ImageBytes)
	if err != nil {
		return component.ResponseError(err)
	}
	err = out.Close()
	if err != nil {
		return component.ResponseError(err)
	}
	return component.ResponseACK
}

func (d *Dummy) Init(config component.Config, logger zap.SugaredLogger) error {
	d.logger = logger
	return nil
}

func (d *Dummy) Start() error {
	d.logger.Info("Started Dummy processor.")
	return nil
}

func (d *Dummy) Close(time.Duration) error {
	return nil
}
