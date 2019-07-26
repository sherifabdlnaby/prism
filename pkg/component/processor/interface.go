package processor

import (
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
)

//------------------------------------------------------------------------------

// Processor can process a payload.
type Processor interface {
	component.Base
	Decoder
	DecoderStream
}

// ReadWrite can decode, process, or encode a payload.
type ReadWrite interface {
	Processor
	ProcessReadWrite
	Encoder
}

// ReadWriteStream can decode, process(R/W), AND encode a payload as a stream.
type ReadWriteStream interface {
	Processor
	ProcessReadWrite
	EncoderStream
}

// ReadOnly can decode, process, and image but doesn't output any data.
type ReadOnly interface {
	Processor
	ProcessReadOnly
}

//------------------------------------------------------------------------------

//Decoder A Base that can decodes an Input
type Decoder interface {
	Decode(in payload.Bytes, data payload.Data) (payload.DecodedImage, response.Response)
}

//DecoderStream A Base that can decodes an Input Stream
type DecoderStream interface {
	DecodeStream(in payload.Stream, data payload.Data) (payload.DecodedImage, response.Response)
}

//------------------------------------------------------------------------------

//Encoder A Base that encode an Image
type Encoder interface {
	Encode(in payload.DecodedImage, data payload.Data) (payload.Bytes, response.Response)
}

//EncoderStream A Base that encode an Image as a stream by writing to supplied outputStream
type EncoderStream interface {
	EncodeStream(in payload.DecodedImage, data payload.Data, out payload.OutputStream) response.Response
}

//------------------------------------------------------------------------------

//ProcessReadWrite A Base that can process an image
type ProcessReadWrite interface {
	Process(in payload.DecodedImage, data payload.Data) (payload.DecodedImage, response.Response)
}

//ProcessReadOnly A base component that can process images in read-only mode (no-output)
type ProcessReadOnly interface {
	Process(in payload.DecodedImage, data payload.Data) response.Response
}
