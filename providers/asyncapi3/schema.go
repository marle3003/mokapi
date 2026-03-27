package asyncapi3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/media"
	openapi "mokapi/providers/openapi/schema"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/encoding"
	jsonSchema "mokapi/schema/json/schema"
	"reflect"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type SchemaRef struct {
	dynamic.Reference
	Value Schema
}

type Schema interface {
	Parse(config *dynamic.Config, reader dynamic.Reader) error
}

type MultiSchemaFormat struct {
	Format string     `yaml:"schemaFormat,omitempty" json:"schemaFormat,omitempty"`
	Schema *SchemaRef `yaml:"schema" json:"schema"`
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
		r.Value = s
	}
	return err
}

func (r *SchemaRef) UnmarshalJSON(b []byte) error {
	d := json.NewDecoder(bytes.NewReader(b))

	err := d.Decode(&r.Reference)
	if err == nil && len(r.Ref) > 0 {
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
		r.Value = s
	}
	return err
}

func (r *SchemaRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		var resolved *SchemaRef
		err := dynamic.Resolve(r.Ref, &resolved, config, reader)
		if err != nil {
			s := &SchemaRef{Value: &jsonSchema.Schema{}}
			err = dynamic.Resolve(r.Ref, &s.Value, config, reader)
			if err != nil {
				return err
			}
			r.Value = s.Value
		} else {
			r.Value = resolved.Value
		}
		return nil
	}
	return r.Value.Parse(config, reader)
}

func (r *SchemaRef) ConvertTo(i interface{}) (interface{}, error) {
	if s, ok := r.Value.(*jsonSchema.Schema); ok {
		return s, nil
	}
	return nil, fmt.Errorf("unsupported schema convert %T: %T", r.Value, i)
}

func (r *SchemaRef) Marshal(v any, ct media.ContentType) ([]byte, error) {
	switch s := r.Value.(type) {
	case *jsonSchema.Schema:
		e := encoding.NewEncoder(s)
		return e.Write(v, ct)
	case *avro.Schema:
		return s.Marshal(v)
	case *openapi.Schema:
		return s.Marshal(v, ct)
	case *SchemaRef:
		return s.Marshal(v, ct)
	case *MultiSchemaFormat:
		return s.Schema.Marshal(v, ct)
	default:
		return nil, fmt.Errorf("unsupported schema type: %T", v)
	}
}

func (r *SchemaRef) GetSchema() (Schema, error) {
	if r.Value == nil {
		return nil, nil
	}

	switch s := r.Value.(type) {
	case *jsonSchema.Schema, *avro.Schema, *openapi.Schema:
		return s, nil
	case *SchemaRef:
		return r.Value, nil
	case *MultiSchemaFormat:
		return s.Schema.GetSchema()
	default:
		return nil, fmt.Errorf("unsupported schema type: %T", s)
	}
}

func (m *MultiSchemaFormat) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if m == nil {
		return nil
	}
	if m.Schema != nil {
		return m.Schema.Parse(config, reader)
	}
	return nil
}

func (m *MultiSchemaFormat) ConvertTo(i interface{}) (interface{}, error) {
	switch i.(type) {
	case *jsonSchema.Schema:
		switch s := m.Schema.Value.(type) {
		case *jsonSchema.Schema:
			return m.Schema, nil
		case *openapi.Schema:
			return openapi.ConvertToJsonSchema(s), nil
		case *avro.Schema:
			return avro.ConvertToJsonSchema(s), nil
		}
	case *openapi.Schema:
		if _, ok := m.Schema.Value.(*openapi.Schema); ok {
			return m.Schema, nil
		}
	case *avro.Schema:
		if _, ok := m.Schema.Value.(*avro.Schema); ok {
			return m.Schema, nil
		}
	}
	return nil, fmt.Errorf("unsupported schema convert %T: %T", m.Schema, i)
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

	ref := &SchemaRef{}
	err = json.Unmarshal(raw, &ref)
	if err == nil && ref.Ref != "" {
		m.Schema = ref
		return nil
	}

	var s Schema
	s, err = unmarshal(raw, m.Format)
	if err != nil {
		return err
	}
	if s != nil {
		m.Schema = &SchemaRef{Value: s}
	}
	return nil
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

	if schemaNode.Kind != yaml.MappingNode {
		return fmt.Errorf("unexpected yaml node kind: %v", node.Kind)
	}

	ref := &SchemaRef{}
	err := schemaNode.Decode(&ref)
	if err == nil && ref.Ref != "" {
		m.Schema = ref
		return nil
	}

	switch {
	case isOpenApi(format):
		var s *openapi.Schema
		err = schemaNode.Decode(&s)
		if err != nil {
			return err
		}
		m.Schema = &SchemaRef{Value: s}
	case isAvro(format):
		var ref *AvroRef
		err = schemaNode.Decode(&ref)
		if err != nil {
			return err
		}
		m.Schema = &SchemaRef{Value: ref}
	default:
		var s *jsonSchema.Schema
		err = schemaNode.Decode(&s)
		if err != nil {
			return err
		}
		m.Schema = &SchemaRef{Value: s}
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
		v1 := reflect.ValueOf(r.Value)
		v2 := reflect.ValueOf(patch.Value)

		// if patch has different schema type then simple overwrite
		if v1.Type() != v2.Type() {
			r.Value = patch.Value
		} else {
			switch s := r.Value.(type) {
			case *openapi.Schema:
				p, ok := patch.Value.(*openapi.Schema)
				if !ok {
					log.Errorf("unexpected patch type: %T", patch.Value)
				} else {
					s.Patch(p)
				}
			case *avro.Schema:
				log.Errorf("patch not supported for Avro schema")
			case *jsonSchema.Schema:
				p, ok := patch.Value.(*jsonSchema.Schema)
				if !ok {
					log.Errorf("unexpected patch type: %T", patch.Value)
				} else {
					s.Patch(p)
				}
			}
		}
	}
}

func (m *MultiSchemaFormat) Patch(patch *MultiSchemaFormat) {
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
		m.Schema.Patch(patch.Schema)
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
