package dummy

import (
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
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
func (d *Dummy) Decode(in transaction.Payload, data transaction.ImageData) (interface{}, response.Response) {
	var imgBytes []byte
	var err error

	// Previous plugins has passed its output as a whole via byte slice,
	// (no need to read it using reader (which will duplicate data in ram) if its available as reference)
	if in.ImageBytes != nil {
		imgBytes = in.ImageBytes
	} else {
		// Read all bytes using input's reader.
		imgBytes, err = ioutil.ReadAll(in)
		if err != nil {
			return transaction.DecodedPayload{}, response.Error(err)
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
func (d *Dummy) Process(dp interface{}, data transaction.ImageData) (interface{}, response.Response) {
	//literally do nothing lol
	time.Sleep(300 + time.Duration(rand.Intn(1500))*time.Millisecond)
	return dp, response.ACK
}

//Encode Simulate Encoding the Image
func (d *Dummy) Encode(in interface{}, data transaction.ImageData, out *transaction.OutputPayload) response.Response {

	// Since in this dummy case we have processed output as a whole, we can just pass it to next node.
	out.ImageBytes = in.(internalImage).internal

	// Write plugin's output, to the output object.
	_, err := out.Write(out.ImageBytes)
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

//startMux startMux the plugin to begin receiving input
func (d *Dummy) Start() error {
	d.logger.Info("Started Dummy processor.")
	return nil
}

//Close Close plugin gracefully
func (d *Dummy) Close() error {
	return nil
}
