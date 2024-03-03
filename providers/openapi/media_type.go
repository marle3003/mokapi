package openapi

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/decoding"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
)

type MediaType struct {
	Schema   *schema.Ref `yaml:"schema,omitempty" json:"schema,omitempty"`
	Example  interface{} `yaml:"example,omitempty" json:"example,omitempty"`
	Examples Examples    `yaml:"examples,omitempty" json:"examples,omitempty"`

	ContentType media.ContentType    `yaml:"-" json:"-"`
	Encoding    map[string]*Encoding `yaml:"encoding,omitempty" json:"encoding,omitempty"`
}

func (m *MediaType) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if m == nil {
		return nil
	}
	if err := m.Schema.Parse(config, reader); err != nil {
		return err
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

	if patch.Example != nil {
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
	var decoder decoding.DecodeFunc
	if contentType.String() == "application/x-www-form-urlencoded" {
		decoder = urlValueDecoder{mt: m}.decode
	}

	v, err := decoding.Decode(b, contentType, decoder)
	if err != nil {
		return nil, err
	}
	p := getParser(contentType)
	return p.Parse(v, m.Schema)
}
