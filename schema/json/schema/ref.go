package schema

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/schema/json/ref"
)

type Ref struct {
	ref.Reference
	Boolean *bool
	Value   *Schema
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
	if r == nil || r.Value == nil {
		return ""
	}
	return fmt.Sprintf("%s", r.Value.Type)
}

func (r *Ref) String() string {
	if r == nil || r.Value == nil {
		return "empty schema"
	}
	return r.Value.String()
}

func (r *Ref) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}
	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}

	if r.Value == nil {
		return nil
	}

	return r.Value.Parse(config, reader)
}

func (r *Ref) IsFreeForm() bool {
	if r == nil {
		return true
	}
	if r.Boolean != nil {
		return *r.Boolean
	}
	return r.Value.IsFreeForm()
}

func (r *Ref) UnmarshalJSON(b []byte) error {
	var boolVal bool
	if err := json.Unmarshal(b, &boolVal); err == nil {
		r.Boolean = &boolVal
		return nil
	}

	return r.UnmarshalJson(b, &r.Value)
}

func (r *Ref) UnmarshalYAML(node *yaml.Node) error {
	var boolVal bool
	if err := node.Decode(&boolVal); err == nil {
		r.Boolean = &boolVal
		return nil
	}

	return r.UnmarshalYaml(node, &r.Value)
}

func NewRef(b bool) *Ref {
	return &Ref{Boolean: &b}
}

func (r *Ref) IsFalse() bool {
	if r == nil {
		return false
	}
	if r.Boolean != nil {
		return !*r.Boolean
	}
	return r.Value.IsFalse()
}
