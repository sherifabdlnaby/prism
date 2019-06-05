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

// ReadWrite can decode, process, or encode a payload.
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
	Decode(in payload.Payload, data payload.Data) (payload.Decoded, response.Response)
}

//Decoder A Component that can decodes an Input
type DecoderStream interface {
	DecodeStream(in payload.Stream, data payload.Data) (payload.Decoded, response.Response)
}

//------------------------------------------------------------------------------

//Encoder A Component that encode an Image
type Encoder interface {
	Encode(in payload.Decoded, data payload.Data) (payload.Payload, response.Response)
}

//Encoder A Component that encode an Image
type EncoderStream interface {
	EncodeStream(in payload.Decoded, data payload.Data, out *payload.Output) response.Response
}

//------------------------------------------------------------------------------

//ProcessReadWrite A Component that can process an image
type ProcessReadWrite interface {
	Process(in payload.Decoded, data payload.Data) (payload.Decoded, response.Response)
}

//Read A base component that can process images in read-only mode (no-output)
type ProcessReadOnly interface {
	Process(in payload.Decoded, data payload.Data) response.Response
}
