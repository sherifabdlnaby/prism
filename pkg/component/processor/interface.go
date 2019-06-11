package processor

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
)

//------------------------------------------------------------------------------

// Base can process a payload.
type Base interface {
	component.Component
	Decoder
	DecoderStream
}

// ReadWrite can decode, process, or encode a payload.
type ReadWrite interface {
	Base
	ProcessReadWrite
	Encoder
}

// ReadWriteStream can decode, process(R/W), AND encode a payload as a stream.
type ReadWriteStream interface {
	Base
	ProcessReadWrite
	EncoderStream
}

// ReadOnly can decode, process, and image but doesn't output any data.
type ReadOnly interface {
	Base
	ProcessReadOnly
}

//------------------------------------------------------------------------------

//Decoder A Component that can decodes an Input
type Decoder interface {
	Decode(in payload.Bytes, data payload.Data) (payload.DecodedImage, response.Response)
}

//DecoderStream A Component that can decodes an Input Stream
type DecoderStream interface {
	DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response)
}

//------------------------------------------------------------------------------

//Encoder A Component that encode an Image
type Encoder interface {
	Encode(in payload.DecodedImage, data payload.Data) (payload.Bytes, response.Response)
}

//EncoderStream A Component that encode an Image as a stream by writing to supplied outputStream
type EncoderStream interface {
	EncodeStream(in payload.DecodedImage, data payload.Data, out payload.OutputStream) response.Response
}

//------------------------------------------------------------------------------

//ProcessReadWrite A Component that can process an image
type ProcessReadWrite interface {
	Process(in payload.DecodedImage, data payload.Data) (payload.DecodedImage, response.Response)
}

//ProcessReadOnly A base component that can process images in read-only mode (no-output)
type ProcessReadOnly interface {
	Process(in payload.DecodedImage, data payload.Data) response.Response
}