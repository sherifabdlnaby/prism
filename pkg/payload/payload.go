package payload

import (
	"io"
)

// Bytes is a byte slice holding the actual bytes of the image.
type Bytes []byte

// Data is a map hold data/metadata about an image that will be read and augmented through the pipeline.
type Data map[string]interface{}

// ---------------------------------------------------------------------------------------------

// Payload holds a reader to image bytes OR a byte slice of the image.
type Stream struct {
	io.Reader
}

// Payload holds a reader to image bytes OR a byte slice of the image.
type Payload struct {
	Bytes
}

// Decoded holds the Image bytes and accompanying Data
type Decoded struct {
	Image interface{}
}

// ---------------------------------------------------------------------------------------------

// Output can either be used to write the image asynchronously OR just pass the whole Bytes.
type Output struct {
	Bytes
}

// Output can either be used to write the image asynchronously OR just pass the whole Bytes.
type OutputStream struct {
	io.WriteCloser
}
