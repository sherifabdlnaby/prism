package vips

import (
	"github.com/h2non/bimg"
	"github.com/sherifabdlnaby/prism/pkg/payload"
)

type export struct {
	Raw exportRawConfig `mapstructure:",squash"`
}

type exportRawConfig struct {
	Format        string `validate:"oneof=jpg jpeg png webp"`
	Extend        string `validate:"oneof=black copy repeat mirror white last"`
	Quality       int    `validate:"min=1,max=100"`
	Compression   int    `validate:"min=1,max=9"`
	StripMetadata bool   `mapstructure:"strip_metadata"`
}

func (o *export) Init() (bool, error) {
	//no need to validate, all validation are done by tags. and no need to create selectors
	return true, nil
}

func (o *export) Apply(p *bimg.Options, data payload.Data) error {

	p.Quality = o.Raw.Quality
	p.Compression = o.Raw.Compression
	p.StripMetadata = o.Raw.StripMetadata

	switch o.Raw.Format {
	case "jpeg", "jpg":
		p.Type = bimg.JPEG
	case "png":
		p.Type = bimg.PNG
	case "webp":
		p.Type = bimg.WEBP
	}

	switch o.Raw.Extend {
	case "black":
		p.Extend = bimg.ExtendBlack
	case "copy":
		p.Extend = bimg.ExtendCopy
	case "repeat":
		p.Extend = bimg.ExtendRepeat
	case "mirror":
		p.Extend = bimg.ExtendMirror
	case "white":
		p.Extend = bimg.ExtendWhite
	case "last":
		p.Extend = bimg.ExtendLast
	}

	return nil
}
