package asyncapi3

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	openapi "mokapi/providers/openapi/schema"
	avro "mokapi/schema/avro/schema"
	json "mokapi/schema/json/schema"
	"reflect"
)

type SchemaRef struct {
	dynamic.Reference
	Value *MultiSchemaFormat
}

type Schema interface {
	Parse(config *dynamic.Config, reader dynamic.Reader) error
}

type MultiSchemaFormat struct {
	Format string `yaml:"schemaFormat" json:"schemaFormat"`
	Schema Schema `yaml:"schema" json:"schema"`
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

	var s *json.Ref
	err = node.Decode(&s)
	if err == nil {
		r.Value = &MultiSchemaFormat{Schema: s}
	}
	return err
}

func (r *SchemaRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *SchemaRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		err := dynamic.Resolve(r.Ref, &r.Value, config, reader)
		if err != nil {
			type t struct {
				s *json.Schema
			}
			s := &t{}
			err = dynamic.Resolve(r.Ref, &s.s, config, reader)
			if err != nil {
				return err
			}
			r.Value = &MultiSchemaFormat{Schema: &json.Ref{Value: s.s}}
		}
	}

	return r.Value.parse(config, reader)
}

func (m *MultiSchemaFormat) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if m.Schema != nil {
		return m.Schema.Parse(config, reader)
	}
	return nil
}

func (m *MultiSchemaFormat) Resolve(token string) (interface{}, error) {
	if token == "" {
		if js, ok := m.Schema.(*json.Ref); ok {
			return js.Value, nil
		}
		return m.Schema, nil
	}
	return m, nil
}

func (m *MultiSchemaFormat) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("unexpected yaml node kind: %v", node.Kind)
	}

	format := ""
	var schemaNode *yaml.Node
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == "schemaFormat" {
			format = node.Content[i+1].Value
		} else if node.Content[i].Value == "schema" {
			schemaNode = node.Content[i+1]
		}
	}

	m.Format = format
	switch format {
	case "application/vnd.oai.openapi;version=3.0.0", "application/vnd.oai.openapi+json;version=3.0.0", "application/vnd.oai.openapi+yaml;version=3.0.0":
		var s *openapi.Ref
		err := schemaNode.Decode(&s)
		if err != nil {
			return err
		}
		m.Schema = s
	case "application/vnd.apache.avro;version=1.9.0", "application/vnd.apache.avro+json;version=1.9.0", "application/vnd.apache.avro+yaml;version=1.9.0":
		var s *avro.Schema
		err := schemaNode.Decode(&s)
		if err != nil {
			return err
		}
		m.Schema = s
	case "":
		var s *json.Ref
		err := node.Decode(&s)
		if err != nil {
			return err
		}
		m.Schema = s
	}

	return nil
}

func (r *SchemaRef) patch(patch *SchemaRef) {
	if r == nil || patch == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value
	} else {
		r.Value.patch(patch.Value)
	}
}

func (m *MultiSchemaFormat) patch(patch *MultiSchemaFormat) {
	if patch == nil {
		return
	}
	if patch.Format != "" {
		m.Format = patch.Format
	}

	if patch.Schema == nil {
		return
	}
	if m.Schema == nil {
		m.Schema = patch.Schema
	} else {
		v1 := reflect.ValueOf(m.Schema)
		v2 := reflect.ValueOf(patch.Schema)

		// if patch has different schema type then simple overwrite
		if v1.Type() != v2.Type() {
			m.Schema = patch.Schema
		} else {
			switch s := m.Schema.(type) {
			case *avro.Schema:
			case *json.Ref:
				s.Patch(patch.Schema.(*json.Ref))
			}
		}
	}
}
