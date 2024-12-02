package asyncapi3

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/schema/json/schema"
)

type SchemaRef struct {
	dynamic.Reference
	Value Schema
}

type Schema interface {
	Parse(config *dynamic.Config, reader dynamic.Reader) error
}

type MultiSchemaFormat struct {
	Format string      `yaml:"schemaFormat" json:"schemaFormat"`
	Schema interface{} `yaml:"schema" json:"schema"`
}

func (r *SchemaRef) UnmarshalYAML(node *yaml.Node) error {
	err := node.Decode(&r.Reference)
	if err == nil && len(r.Ref) > 0 {
		return nil
	}

	var multi *MultiSchemaFormat
	err = node.Decode(&multi)
	if err == nil && multi.Format != "" {
		r.Value = multi
		return nil
	}

	var s *schema.Schema
	err = node.Decode(&s)
	if err == nil {
		r.Value = s
	}
	return err
}

func (r *SchemaRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *SchemaRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}

	return r.Value.Parse(config, reader)
}

func (r *MultiSchemaFormat) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	return nil
}

func ConvertToJsonSchema(s Schema) *schema.Schema {
	if js, ok := s.(*schema.Schema); ok {
		return js
	}
	panic(fmt.Sprintf("unknown schema: %T", s))
}
