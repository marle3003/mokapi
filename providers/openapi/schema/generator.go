package schema

import (
	"mokapi/schema/json/generator"
)

func CreateValue(s *Schema) (interface{}, error) {
	if s == nil {
		return generator.New(&generator.Request{})
	}
	c := &JsonSchemaConverter{}
	r := &generator.Request{Schema: c.Convert(s)}
	if s.Xml != nil && s.Xml.Name != "" {
		r.Path = append(r.Path, s.Xml.Name)
	}

	return generator.New(r)
}
