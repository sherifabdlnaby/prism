package payload

import (
	"io"
)

// Data is a map hold data/metadata about an image that will be read and augmented through the pipeline.
type Data map[string]interface{}

// ---------------------------------------------------------------------------------------------

// Stream holds a reader to image bytes
type Stream io.Reader

// Bytes a byte slice of an image file.
type Bytes []byte

//Payload is either a payload.Stream OR payload.Bytes
type Payload interface{}

// DecodedImage holds the Image bytes and accompanying Data
type DecodedImage interface{}

// ---------------------------------------------------------------------------------------------

// Output a byte slice of an encoded image file.
type Output []byte

// OutputStream can either be used to write the image asynchronously.
type OutputStream io.WriteCloser
