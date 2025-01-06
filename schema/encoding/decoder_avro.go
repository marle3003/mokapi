package encoding

import (
	"mokapi/media"
)

type AvroDecoder struct {
}

func (d *AvroDecoder) IsSupporting(contentType media.ContentType) bool {
	return contentType.String() == "avro/binary"
}

func (d *AvroDecoder) Decode(b []byte, state *DecodeState) (i interface{}, err error) {
	return state.parser.Parse(b)
}
