package schema

import (
	"io"
	"mokapi/media"
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
)

func ParseString(s string, schema *Ref) (interface{}, error) {
	return encoding.DecodeString(
		s,
		encoding.WithSchema(ConvertToJsonSchema(schema)),
		encoding.WithParser(&parser.Parser{ConvertStringToNumber: true, ConvertStringToBoolean: true}))
}

func UnmarshalFrom(r io.Reader, contentType media.ContentType, schema *Ref) (i interface{}, err error) {
	if contentType.IsXml() {
		return UnmarshalXML(r, schema)
	}

	return encoding.DecodeFrom(r, encoding.WithContentType(contentType), encoding.WithSchema(ConvertToJsonSchema(schema)))
}
