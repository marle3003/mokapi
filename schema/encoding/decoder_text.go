package encoding

import (
	"mokapi/media"
)

type TextDecoder struct {
}

func (d *TextDecoder) IsSupporting(contentType media.ContentType) bool {
	return contentType.Type == "text"
}

func (d *TextDecoder) Decode(b []byte, state *DecodeState) (i interface{}, err error) {
	return state.parser.Parse(string(b))
}
