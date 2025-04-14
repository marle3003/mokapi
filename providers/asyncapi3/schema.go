package asyncapi3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	openapi "mokapi/providers/openapi/schema"
	avro "mokapi/schema/avro/schema"
	jsonSchema "mokapi/schema/json/schema"
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
	Format string `yaml:"schemaFormat,omitempty" json:"schemaFormat,omitempty"`
	Schema Schema `yaml:"schema" json:"schema"`
}

func (r *SchemaRef) UnmarshalYAML(node *yaml.Node) error {
	err := node.Decode(&r.Reference)
	if err == nil && len(r.Ref) > 0 {
		return nil
	}

	var multi *MultiSchemaFormat
	err = node.Decode(&multi)
	if err == nil && multi.Format != "" || multi.Schema != nil {
		r.Value = multi
		return nil
	}

	var s *jsonSchema.Schema
	err = node.Decode(&s)
	if err == nil {
		r.Value = &MultiSchemaFormat{Schema: s}
	}
	return err
}

func (r *SchemaRef) UnmarshalJSON(b []byte) error {
	d := json.NewDecoder(bytes.NewReader(b))

	err := d.Decode(&r.Ref)
	if err == nil {
		return nil
	}

	d = json.NewDecoder(bytes.NewReader(b))
	var multi *MultiSchemaFormat
	err = d.Decode(&multi)
	if err == nil && multi.Format != "" || multi.Schema != nil {
		r.Value = multi
		return nil
	}

	d = json.NewDecoder(bytes.NewReader(b))
	var s *jsonSchema.Schema
	err = d.Decode(&s)
	if err == nil {
		r.Value = &MultiSchemaFormat{Schema: s}
	}
	return err
}

func (r *SchemaRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		err := dynamic.Resolve(r.Ref, &r.Value, config, reader)
		if err != nil {
			type t struct {
				s *jsonSchema.Schema
			}
			s := &t{}
			err = dynamic.Resolve(r.Ref, &s.s, config, reader)
			if err != nil {
				return err
			}
			r.Value = &MultiSchemaFormat{Schema: s.s}
		}
		return nil
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
		if js, ok := m.Schema.(*jsonSchema.Schema); ok {
			return js, nil
		}
		return m.Schema, nil
	}
	return m, nil
}

func (m *MultiSchemaFormat) UnmarshalJSON(b []byte) error {
	d := json.NewDecoder(bytes.NewReader(b))

	token, err := d.Token()
	if err != nil {
		return err
	}

	if delim, ok := token.(json.Delim); ok && delim != '{' {
		return fmt.Errorf("unexpected token %s; expected '{'", token)
	}

	var raw json.RawMessage
	for {
		token, err = d.Token()
		if err != nil {
			return err
		}

		if delim, ok := token.(json.Delim); ok && delim == '}' {
			break
		}

		switch token.(string) {
		case "schemaFormat":
			token, err = d.Token()
			if err != nil {
				return err
			}
			m.Format = token.(string)
		case "schema":
			err = d.Decode(&raw)
			if err != nil {
				return err
			}
		}
	}

	m.Schema, err = unmarshal(raw, m.Format)

	return err
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

	if schemaNode == nil {
		return nil
	}

	switch {
	case isOpenApi(format):
		var s *openapi.Schema
		err := schemaNode.Decode(&s)
		if err != nil {
			return err
		}
		m.Schema = s
	case isAvro(format):
		var ref *AvroRef
		err := schemaNode.Decode(&ref)
		if err != nil {
			return err
		}
		m.Schema = ref
	default:
		var s *jsonSchema.Schema
		err := schemaNode.Decode(&s)
		if err != nil {
			return err
		}
		m.Schema = s
	}

	return nil
}

func (r *SchemaRef) Patch(patch *SchemaRef) {
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
			case *jsonSchema.Schema:
				s.Patch(patch.Schema.(*jsonSchema.Schema))
			}
		}
	}
}

func unmarshal(raw json.RawMessage, format string) (Schema, error) {
	if raw != nil {
		switch {
		case isOpenApi(format):
			var r *openapi.Schema
			err := json.Unmarshal(raw, &r)
			return r, err
		case isAvro(format):
			var a *avro.Schema
			err := json.Unmarshal(raw, &a)
			return a, err
		default:
			var r *jsonSchema.Schema
			err := json.Unmarshal(raw, &r)
			return r, err
		}
	}
	return nil, nil
}

func isAvro(format string) bool {
	switch format {
	case "application/vnd.apache.avro;version=1.9.0",
		"application/vnd.apache.avro+json;version=1.9.0":
		return true
	default:
		return false
	}
}

func isOpenApi(format string) bool {
	switch format {
	case "application/vnd.oai.openapi+json;version=3.0.0",
		"application/vnd.oai.openapi;version=3.0.0":
		return true
	default:
		return false
	}
}

type AvroRef struct {
	*avro.Schema
	dynamic.Reference
}

func (r *AvroRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r.Ref != "" {
		return dynamic.Resolve(r.Ref, &r.Schema, config, reader)
	}
	return nil
}

func (r *AvroRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Schema)
}

func (r *AvroRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Schema)
}
