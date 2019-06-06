package payload

import (
	"io"
)

// Data is a map hold data/metadata about an image that will be read and augmented through the pipeline.
type Data map[string]interface{}

// ---------------------------------------------------------------------------------------------

// Payload holds a reader to image bytes OR a byte slice of the image.
type Stream io.Reader

// Payload holds a reader to image bytes OR a byte slice of the image.
type Payload []byte

// DecodedImage holds the Image bytes and accompanying Data
type DecodedImage interface{}

// ---------------------------------------------------------------------------------------------

// Output can either be used to write the image asynchronously OR just pass the whole Bytes.
type Output []byte

// Output can either be used to write the image asynchronously OR just pass the whole Bytes.
type OutputStream io.WriteCloser
