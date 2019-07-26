package dummy

import (
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"go.uber.org/zap"
)

//Dummy Dummy ProcessReadWrite that does absolutely nothing to the image
type Dummy struct {
	logger zap.SugaredLogger
}

type internalImage struct {
	internal []byte
}

// NewComponent Return a new Base
func NewComponent() component.Base {
	return &Dummy{}
}

//Decode Simulate Decoding the Image
func (d *Dummy) Decode(in payload.Bytes, data payload.Data) (payload.DecodedImage, response.Response) {
	// create internal decoded object (varies with each plugin)`
	out := internalImage{
		internal: in,
	}

	// Return it as it is (dummy).
	return out, response.ACK
}

//DecodeStream Simulate Decoding the Image
func (d *Dummy) DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response) {
	var imgBytes []byte
	var err error

	imgBytes, err = ioutil.ReadAll(in)
	if err != nil {
		return nil, response.Error(err)
	}

	// create internal decoded object (varies with each plugin)`
	out := internalImage{
		internal: imgBytes,
	}

	// Return it as it is (dummy).
	return out, response.ACK
}

//Process Simulate Processing the Image
func (d *Dummy) Process(in payload.DecodedImage, data payload.Data) (payload.DecodedImage, response.Response) {
	//literally do nothing lol
	time.Sleep(1000 + time.Duration(rand.Intn(1500))*time.Millisecond)
	return in, response.ACK
}

//Encode Simulate Encoding the Image
func (d *Dummy) Encode(in payload.DecodedImage, data payload.Data) (payload.Bytes, response.Response) {
	// Since in this dummy case we have processed output as a whole, we can just pass it to next node.
	Payload := in.(internalImage).internal

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

//Stop Stop plugin gracefully
func (d *Dummy) Stop() error {
	return nil
}
