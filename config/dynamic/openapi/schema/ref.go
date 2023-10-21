package schema

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/ref"
)

type Ref struct {
	ref.Reference
	Value *Schema
}

func (r *Ref) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}
	if len(r.Ref) > 0 {
		if err := common.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return fmt.Errorf("parse schema failed: %w", err)
		}
		return nil
	}

	if r.Value == nil {
		return nil
	}

	return r.Value.Parse(config, reader)
}

func (r *Ref) UnmarshalYAML(node *yaml.Node) error {
	return r.Unmarshal(node, &r.Value)
}

func (r *Ref) UnmarshalJSON(b []byte) error {
	return r.UnmarshalJson(b, &r.Value)
}

func (r *Ref) HasProperties() bool {
	return r.Value != nil && r.Value.HasProperties()
}

func (r *Ref) getXml() *Xml {
	if r != nil && r.Value != nil {
		return r.Value.Xml
	}
	return nil
}

func (r *Ref) getProperty(name string) *Ref {
	if r == nil && r.Value == nil {
		return nil
	}
	return r.Value.Properties.Get(name)
}

func (r *Ref) getPropertyXml(name string) *Xml {
	prop := r.getProperty(name)
	if prop == nil {
		return nil
	}
	return prop.getXml()
}
