package dummy

import (
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"io"
	"time"
)

type Dummy struct {
	logger zap.SugaredLogger
}

func NewComponent() types.Component {
	return &Dummy{}
}

func (d *Dummy) Decode(ep types.Payload) (types.DecodedPayload, error) {
	d.logger.Infow("Decoding Payload... ", "name", ep.Name)

	// Return it as it is (dummy).
	return types.DecodedPayload{
		Name:      "test",
		Image:     ep.Reader,
		ImageData: nil,
	}, nil
}

func (d *Dummy) Process(dp types.DecodedPayload) (types.DecodedPayload, error) {
	d.logger.Infow("Processing Payload... ", "name", dp.Name)
	return dp, nil
}

func (d *Dummy) Encode(dp types.DecodedPayload) (types.Payload, error) {

	d.logger.Infow("Encoding Payload... ", "name", dp.Name)

	return types.Payload{
		Name:      "test",
		Reader:    dp.Image.(io.Reader),
		ImageData: nil,
	}, nil
}

func (d *Dummy) Init(config types.Config, logger zap.SugaredLogger) error {
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
