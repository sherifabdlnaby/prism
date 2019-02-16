package processor

import (
	"github.com/sherifabdlnaby/prism/lib/pkg/types"
	"log"
	"time"
)

type Dummy struct {
}

func (*Dummy) Decode(ep types.EncodedPayload) (types.DecodedPayload, error) {
	log.Println("Decoding Payload... ", ep.Name)

	// Return it as it is.
	return types.DecodedPayload{
		Name:      "test",
		Image:     ep.ImageBytes,
		ImageData: nil,
	}, nil
}

func (*Dummy) Process(dp types.DecodedPayload) (types.DecodedPayload, error) {
	log.Println("Processing Payload... ", dp.Name)
	return dp, nil
}

func (*Dummy) Encode(dp types.DecodedPayload) (types.EncodedPayload, error) {

	log.Println("Encoding Payload... ", dp.Name)

	return types.EncodedPayload{
		Name:       "test",
		ImageBytes: dp.Image.(types.ImageBytes),
		ImageData:  nil,
	}, nil
}

func (*Dummy) Init(types.Config) error {
	log.Println("Initialized Dummy processor.")
	return nil
}

func (*Dummy) Start() error {
	log.Println("Started Dummy processor.")
	return nil
}

func (*Dummy) Close(time.Duration) error {
	return nil
}
