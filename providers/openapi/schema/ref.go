package schema

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type Ref struct {
	dynamic.Reference
	Value *Schema
}

func (r *Ref) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
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
	return r.UnmarshalYaml(node, &r.Value)
}

func (r *Ref) UnmarshalJSON(b []byte) error {
	return r.UnmarshalJson(b, &r.Value)
}

func (r *Ref) HasProperties() bool {
	return r.Value != nil && r.Value.HasProperties()
}

func (r *Ref) String() string {
	if r.Value == nil && len(r.Ref) == 0 {
		return fmt.Sprintf("no schema defined")
	}
	if r.Value == nil {
		return fmt.Sprintf("unresolved schema %v", r.Ref)
	}
	return r.Value.String()
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

func (r *Ref) IsXmlWrapped() bool {
	return r.Value != nil && r.Value.Xml != nil && r.Value.Xml.Wrapped
}
