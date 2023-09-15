package openapi

import (
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/media"
)

type MediaType struct {
	Schema   *schema.Ref `yaml:"schema,omitempty" json:"schema,omitempty"`
	Example  interface{} `yaml:"example,omitempty" json:"example,omitempty"`
	Examples Examples    `yaml:"examples,omitempty" json:"examples,omitempty"`

	ContentType media.ContentType `yaml:"-" json:"-"`
}

func (m *MediaType) parse(config *common.Config, reader common.Reader) error {
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
