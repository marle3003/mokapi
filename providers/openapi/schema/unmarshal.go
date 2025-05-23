package schema

import (
	"mokapi/schema/json/parser"
)

func ParseString(s string, schema *Schema) (interface{}, error) {
	p := parser.Parser{Schema: ConvertToJsonSchema(schema), ConvertStringToNumber: true, ConvertStringToBoolean: true}
	return p.Parse(s)
}
