package types

import (
	"io"
)

// ImageBytes is a byte slice holding the actual bytes of the image.
type ImageBytes []byte

// ImageData is a map hold data/metadata about an image that will be read and augmented through the pipeline.
type ImageData map[string]interface{}

// Payload holds the Image bytes and accompanying Data
type Payload struct {
	Name string
	ImageBytes
	ImageData
}

// StreamablePayload holds a reader to image bytes and accompanying Data.
type StreamablePayload struct {
	Name string
	io.Reader
	ImageData
}
