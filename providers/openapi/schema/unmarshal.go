package schema

import (
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
)

func ParseString(s string, schema *Ref) (interface{}, error) {
	return encoding.DecodeString(
		s,
		encoding.WithParser(&parser.Parser{Schema: ConvertToJsonSchema(schema), ConvertStringToNumber: true, ConvertStringToBoolean: true}))
}
