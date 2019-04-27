package vips

import (
	"io/ioutil"
	"runtime"

	"github.com/sherifabdlnaby/govips/pkg/vips"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/config"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

//Dummy Dummy Processor that does absolutely nothing to the image
type Vips struct {
	logger     zap.SugaredLogger
	operations []func(ref *vips.ImageRef, data transaction.ImageData) error
}

type internalImage struct {
	internal *vips.ImageRef
}

func init() {
	vips.Startup(&vips.Config{
		ConcurrencyLevel: 4,
	})
}

// NewComponent Return a new Component
func NewComponent() component.Component {
	return &Vips{}
}

//Decode Simulate Decoding the Image
func (d *Vips) Decode(in transaction.Payload, data transaction.ImageData) (interface{}, response.Response) {

	buff, _ := ioutil.ReadAll(in.Reader)
	defer runtime.KeepAlive(buff)

	img, err := vips.NewImageFromBuffer(buff)

	if err != nil {
		return transaction.DecodedPayload{}, response.Error(err)
	}

	// create internal object (varies with each plugin)`
	out := internalImage{
		internal: img,
	}

	// Return it as it is (dummy).
	return out, response.ACK
}

//Process Simulate Processing the Image
func (d *Vips) Process(dp interface{}, data transaction.ImageData) (interface{}, response.Response) {
	origImg := dp.(internalImage)
	img := vips.NewImageFromRef(origImg.internal)

	for _, op := range d.operations {
		err := op(img, data)
		if err != nil {
			return nil, response.Error(err)
		}
	}

	return internalImage{
		internal: img,
	}, response.ACK
}

//Encode Simulate Encoding the Image
func (d *Vips) Encode(in interface{}, data transaction.ImageData, out *transaction.OutputPayload) response.Response {
	Img := in.(internalImage)

	byteBuff, _, err := Img.internal.Export(vips.ExportParams{
		Format:  vips.ImageTypeJPEG,
		Quality: 95,
	})
	if err != nil {
		return response.Error(err)
	}

	_, err = out.Write(byteBuff)
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
func (d *Vips) Init(config config.Config, logger zap.SugaredLogger) error {
	d.logger = logger

	// hardcodeit
	d.operations = []func(ref *vips.ImageRef, data transaction.ImageData) error{
		func(ref *vips.ImageRef, data transaction.ImageData) error {
			// Thumbnail
			width := data["width"].(int)
			return ref.ThumbnailImage(width)
		},
		func(ref *vips.ImageRef, data transaction.ImageData) error {
			// Flip
			width := data["flip"].(int)
			return ref.Flip(vips.Direction(width) % 3)
		},
		func(ref *vips.ImageRef, data transaction.ImageData) error {
			// Flip
			width := data["blur"].(float64)
			return ref.Gaussblur(width)
		},
	}

	return nil
}

//Start Start the plugin to begin receiving input
func (d *Vips) Start() error {
	d.logger.Info("Started Dummy processor.")
	return nil
}

//Close Close plugin gracefully
func (d *Vips) Close() error {
	return nil
}
