package encoding

import (
	"mokapi/media"
)

type TextDecoder struct {
}

func (d *TextDecoder) IsSupporting(contentType media.ContentType) bool {
	return contentType.Type == "text"
}

func (d *TextDecoder) Decode(b []byte, _ media.ContentType, _ DecodeFunc) (i interface{}, err error) {
	return string(b), nil
}
