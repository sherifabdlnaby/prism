package dummy

import (
	"github.com/sherifabdlnaby/prism/pkg/types"
	"go.uber.org/zap"
	"io/ioutil"
	"time"
)

type Dummy struct {
	logger zap.SugaredLogger
}

func NewComponent() types.Component {
	return &Dummy{}
}

func (d *Dummy) Decode(ep types.InputPayload) (types.DecodedPayload, error) {
	d.logger.Debugw("Decoding InputPayload... ")

	imgBytes, err := ioutil.ReadAll(ep)

	if err != nil {
		return types.DecodedPayload{}, nil
	}

	// Return it as it is (dummy).
	return types.DecodedPayload{
		Image:     imgBytes,
		ImageData: ep.ImageData,
	}, nil
}

func (d *Dummy) Process(dp types.DecodedPayload) (types.DecodedPayload, error) {
	d.logger.Debugw("Processing InputPayload... ")
	return dp, nil
}

func (d *Dummy) Encode(in types.DecodedPayload, out *types.OutputPayload) error {
	d.logger.Debugw("Encoding InputPayload... ")
	out.ImageBytes = in.Image.([]byte)
	_, err := out.Write(out.ImageBytes)
	if err != nil {
		return err
	}
	err = out.Close()
	if err != nil {
		return err
	}
	return nil
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
