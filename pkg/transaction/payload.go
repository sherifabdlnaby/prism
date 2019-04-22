package transaction

import (
	"io"
)

// ImageBytes is a byte slice holding the actual bytes of the image.
type ImageBytes []byte

// ImageData is a map hold data/metadata about an image that will be read and augmented through the pipeline.
type ImageData map[string]interface{}

// DecodedPayload holds the Image bytes and accompanying Data
type DecodedPayload struct {
	Image interface{}
}

// Payload holds a reader to image bytes OR a byte slice of the image.
type Payload struct {
	io.Reader
	ImageBytes
}

// OutputPayload can either be used to write the image asynchronously OR just pass the whole ImageBytes.
type OutputPayload struct {
	io.WriteCloser
	ImageBytes
}
