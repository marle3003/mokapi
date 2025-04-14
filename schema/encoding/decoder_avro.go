package encoding

import (
	"mokapi/media"
)

type AvroDecoder struct {
}

func (d *AvroDecoder) IsSupporting(contentType media.ContentType) bool {
	ct := contentType.String()
	return ct == "avro/binary" || ct == "application/avro" || ct == "application/octet-stream"
}

func (d *AvroDecoder) Decode(b []byte, state *DecodeState) (i interface{}, err error) {
	return state.parser.Parse(b)
}
