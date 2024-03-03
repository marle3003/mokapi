package decoding

import (
	"fmt"
	"mokapi/media"
)

type Decoder interface {
	IsSupporting(contentType media.ContentType) bool
	Decode([]byte, media.ContentType, DecodeFunc) (interface{}, error)
}

type DecodeFunc func(propName string, val interface{}) (interface{}, error)

var decoders []Decoder

func init() {
	RegisterDecoder(&JsonDecoder{})
	RegisterDecoder(&XmlDecoder{})
	RegisterDecoder(&FormUrlEncodeDecoder{})
	RegisterDecoder(&MultipartDecoder{})
	RegisterDecoder(&FileDecoder{})
	RegisterDecoder(&TextDecoder{})
}

func RegisterDecoder(d Decoder) {
	decoders = append(decoders, d)
}

func Decode(b []byte, contentType media.ContentType, decode DecodeFunc) (interface{}, error) {
	for _, d := range decoders {
		if d.IsSupporting(contentType) {
			return d.Decode(b, contentType, decode)
		}
	}
	return nil, fmt.Errorf("unsupported content type: %v", contentType)
}
