package dummy

import (
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/sherifabdlnaby/prism/pkg/component/processor"
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

// NewComponent Return a new Component
func NewComponent() processor.Component {
	return &Dummy{}
}

//Decode Simulate Decoding the Image
func (d *Dummy) Decode(in payload.Payload, data payload.Data) (interface{}, response.Response) {
	var imgBytes []byte
	var err error

	// Previous plugins has passed its output as a whole via byte slice,
	// (no need to read it using reader (which will duplicate data in ram) if its available as reference)
	if in.Bytes != nil {
		imgBytes = in.Bytes
	} else {
		// Read all bytes using input's reader.
		imgBytes, err = ioutil.ReadAll(in)
		if err != nil {
			return payload.Decoded{}, response.Error(err)
		}
	}

	// create internal decoded object (varies with each plugin)`
	out := internalImage{
		internal: imgBytes,
	}

	// Return it as it is (dummy).
	return out, response.ACK
}

//Process Simulate Processing the Image
func (d *Dummy) Process(dp interface{}, data payload.Data) (interface{}, response.Response) {
	//literally do nothing lol
	time.Sleep(50000 + time.Duration(rand.Intn(1500))*time.Millisecond)
	return dp, response.ACK
}

//Encode Simulate Encoding the Image
func (d *Dummy) Encode(in interface{}, data payload.Data, out *payload.Output) response.Response {

	// Since in this dummy case we have processed output as a whole, we can just pass it to next node.
	out.Bytes = in.(internalImage).internal

	// Write plugin's output, to the output object.
	_, err := out.Write(out.Bytes)
	if err != nil {
		return response.Error(err)
	}

	// Close() must be called to indicate EOF.
	err = out.Close()
	if err != nil {
		return response.Error(err)
	}

	return response.ACK
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
