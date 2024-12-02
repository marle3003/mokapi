package schema

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/schema/json/ref"
)

type Ref struct {
	ref.Reference
	Value *Schema
}

func (r *Ref) IsAny() bool {
	return r == nil || r.Value == nil || len(r.Value.Type) == 0
}

func (r *Ref) IsString() bool {
	return r != nil && r.Value != nil && r.Value.IsString()
}

func (r *Ref) IsInteger() bool {
	return r != nil && r.Value != nil && r.Value.IsInteger()
}

func (r *Ref) IsNumber() bool {
	return r != nil && r.Value != nil && r.Value.IsNumber()
}

func (r *Ref) IsArray() bool {
	return r != nil && r.Value != nil && r.Value.IsArray()
}

func (r *Ref) IsObject() bool {
	return r != nil && r.Value != nil && r.Value.IsObject()
}

func (r *Ref) IsNullable() bool {
	return r != nil && r.Value != nil && r.Value.IsNullable()
}

func (r *Ref) IsDictionary() bool {
	return r != nil && r.Value != nil && r.Value.IsDictionary()
}

func (r *Ref) HasProperties() bool {
	return r != nil && r.Value != nil && r.Value.HasProperties()
}

func (r *Ref) IsAnyString() bool {
	return r != nil && r.Value != nil && r.Value.IsAnyString()
}

func (r *Ref) IsOneOf(typeNames ...string) bool {
	return r != nil && r.Value != nil && r.Value.Type.IsOneOf(typeNames...)
}

func (r *Ref) Type() string {
	if r == nil || r.Value == nil || len(r.Value.Type) == 0 {
		return ""
	}
	if len(r.Value.Type) == 1 {
		return r.Value.Type[0]
	}
	return fmt.Sprintf("%v", r.Value.Type)
}

func (r *Ref) String() string {
	if r == nil || r.Value == nil {
		return ""
	}
	return r.Value.String()
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

func (r *Ref) UnmarshalJSON(b []byte) error {
	return r.UnmarshalJson(b, &r.Value)
}

func (r *Ref) UnmarshalYAML(node *yaml.Node) error {
	return r.UnmarshalYaml(node, &r.Value)
}
