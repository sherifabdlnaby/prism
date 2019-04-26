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
	//d.logger.Debugw("Decoding Payload... ")

	imgBytes, err := ioutil.ReadAll(in)

	if err != nil {
		return transaction.DecodedPayload{}, response.Error(err)
	}

	// create internal object (varies with each plugin)`
	out := internalImage{
		internal: imgBytes,
	}

	// Return it as it is (dummy).
	return out, response.ACK
}

//Process Simulate Processing the Image
func (d *Dummy) Process(dp interface{}, data transaction.ImageData) (interface{}, response.Response) {
	//d.logger.Debugw("Processing Payload... ")
	time.Sleep(300 + time.Duration(rand.Intn(1500))*time.Millisecond)
	return dp, response.ACK
}

//Encode Simulate Encoding the Image
func (d *Dummy) Encode(in interface{}, data transaction.ImageData, out *transaction.OutputPayload) response.Response {
	//d.logger.Debugw("Encoding Payload... ")
	out.ImageBytes = in.(internalImage).internal
	_, err := out.Write(out.ImageBytes)
	if err != nil {
		return response.Error(err)
	}
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

//Start Start the plugin to begin receiving input
func (d *Dummy) Start() error {
	d.logger.Info("Started Dummy processor.")
	return nil
}

//Close Close plugin gracefully
func (d *Dummy) Close() error {
	return nil
}
