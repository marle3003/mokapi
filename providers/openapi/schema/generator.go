package schema

import (
	"mokapi/schema/json/generator"
)

func CreateValue(s *Schema) (interface{}, error) {
	if s == nil || s.SubSchema == nil {
		return generator.New(&generator.Request{})
	}
	c := &JsonSchemaConverter{}
	r := c.Convert(s)
	return generator.New(&generator.Request{Path: generator.Path{&generator.PathElement{Schema: r}}})
}
