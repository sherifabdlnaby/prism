package processor

import (
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"time"
)

type Dummy struct {
	logger zap.Logger
}

func (d *Dummy) Decode(ep types.EncodedPayload) (types.DecodedPayload, error) {
	d.logger.Info("Decoding Payload... ", zap.String("name", ep.Name))

	// Return it as it is.
	return types.DecodedPayload{
		Name:      "test",
		Image:     ep.ImageBytes,
		ImageData: nil,
	}, nil
}

func (d *Dummy) Process(dp types.DecodedPayload) (types.DecodedPayload, error) {
	d.logger.Info("Processing Payload... ", zap.String("name", dp.Name))
	return dp, nil
}

func (d *Dummy) Encode(dp types.DecodedPayload) (types.EncodedPayload, error) {

	d.logger.Info("Encoding Payload... ", zap.String("name", dp.Name))

	return types.EncodedPayload{
		Name:       "test",
		ImageBytes: dp.Image.(types.ImageBytes),
		ImageData:  nil,
	}, nil
}

func (d *Dummy) Init(config types.Config, logger zap.Logger) error {
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
