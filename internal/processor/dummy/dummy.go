package dummy

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"go.uber.org/zap"
	"io/ioutil"
	"time"
)

//Dummy Dummy Processor that does absolutely nothing to the image
type Dummy struct {
	logger zap.SugaredLogger
}

type internalImage struct {
	internal []byte
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &Dummy{}
}

//Decode Simulate Decoding the Image
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

//Process Simulate Processing the Image
func (d *Dummy) Process(dp interface{}, data component.ImageData) (interface{}, component.Response) {
	d.logger.Debugw("Processing InputPayload... ")
	return dp, component.ResponseACK
}

//Encode Simulate Encoding the Image
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

//Init Initialize Plugin based on parsed config
func (d *Dummy) Init(config component.Config, logger zap.SugaredLogger) error {
	d.logger = logger
	return nil
}

//Start Start the plugin to begin receiving input
func (d *Dummy) Start() error {
	d.logger.Info("Started Dummy processor.")
	return nil
}

//Close Close plugin gracefully
func (d *Dummy) Close(time.Duration) error {
	return nil
}
