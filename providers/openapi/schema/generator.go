package schema

import (
	"mokapi/schema/json/generator"
)

func CreateValue(ref *Ref) (interface{}, error) {
	if ref == nil {
		return generator.New(&generator.Request{})
	}
	c := &JsonSchemaConverter{}
	r := c.ConvertToJsonRef(ref)
	return generator.New(&generator.Request{Path: generator.Path{&generator.PathElement{Schema: r}}})
}
