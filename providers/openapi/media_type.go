package openapi

import (
	"bytes"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
)

type MediaType struct {
	Schema   *schema.Schema `yaml:"schema,omitempty" json:"schema,omitempty"`
	Example  *ExampleValue  `yaml:"example,omitempty" json:"example,omitempty"`
	Examples Examples       `yaml:"examples,omitempty" json:"examples,omitempty"`

	ContentType media.ContentType    `yaml:"-" json:"-"`
	Encoding    map[string]*Encoding `yaml:"encoding,omitempty" json:"encoding,omitempty"`
}

func (m *MediaType) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if m == nil {
		return nil
	}
	if err := m.Schema.Parse(config, reader); err != nil {
		return fmt.Errorf("parse schema failed: %s", err)
	}

	if err := m.Examples.parse(config, reader); err != nil {
		return err
	}

	return nil
}

func (m *MediaType) patch(patch *MediaType) {
	if m.Schema == nil {
		m.Schema = patch.Schema
	} else {
		m.Schema.Patch(patch.Schema)
	}

	if patch.Example != nil && patch.Example.Value != nil {
		m.Example = patch.Example
		if len(m.Examples) > 0 {
			m.Examples = nil
		}
	}

	if m.Examples == nil && patch.Examples != nil {
		m.Examples = patch.Examples
		m.Example = nil
	} else if m.Examples != nil {
		m.Examples.patch(patch.Examples)
		m.Example = nil
	}
}

func (m *MediaType) Parse(b []byte, contentType media.ContentType) (interface{}, error) {
	if !contentType.IsDerivedFrom(m.ContentType) {
		return nil, fmt.Errorf("content type '%v' does not match: %v", m.ContentType, contentType)
	}

	if contentType.IsXml() {
		return schema.UnmarshalXML(bytes.NewReader(b), m.Schema)
	}

	p := &parser.Parser{Schema: schema.ConvertToJsonSchema(m.Schema), ValidateAdditionalProperties: true}
	opts := []encoding.DecodeOptions{
		encoding.WithContentType(contentType),
		encoding.WithParser(p),
	}

	if contentType.Type == "text" {
		p.ConvertStringToNumber = true
	}
	if contentType.String() == "application/x-www-form-urlencoded" {
		p.ConvertStringToNumber = true
		opts = append(opts, encoding.WithDecodeFormUrlParam(urlValueDecoder{mt: m}.decode))
	}
	if contentType.Key() == "multipart/form-data" {
		opts = append(opts, encoding.WithDecodePart(multipartForm{mt: m}.decode))
	}

	return encoding.Decode(b, opts...)
}
