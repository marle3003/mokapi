package schema

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/schema/json/schema"
	"strings"
)

type Schema struct {
	Id         string `yaml:"$id,omitempty" json:"$id,omitempty"`
	Ref        string `yaml:"$ref,omitempty" json:"$ref,omitempty"`
	DynamicRef string `yaml:"$dynamicRef,omitempty" json:"$dynamicRef,omitempty"`

	Schema        string `yaml:"$schema,omitempty" json:"$schema,omitempty"`
	Boolean       *bool  `yaml:"-" json:"-"`
	Anchor        string `yaml:"$anchor,omitempty" json:"$anchor,omitempty"`
	DynamicAnchor string `yaml:"$dynamicAnchor,omitempty" json:"$dynamicAnchor,omitempty"`

	Type  schema.Types  `yaml:"type,omitempty" json:"type,omitempty"`
	Enum  []interface{} `yaml:"enum,omitempty" json:"enum,omitempty"`
	Const *interface{}  `yaml:"const,omitempty" json:"const,omitempty"`

	// Numbers
	MultipleOf       *float64                         `yaml:"multipleOf,omitempty" json:"multipleOf,omitempty"`
	Minimum          *float64                         `yaml:"minimum,omitempty" json:"minimum,omitempty"`
	Maximum          *float64                         `yaml:"maximum,omitempty" json:"maximum,omitempty"`
	ExclusiveMinimum *schema.UnionType[float64, bool] `yaml:"exclusiveMinimum,omitempty" json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *schema.UnionType[float64, bool] `yaml:"exclusiveMaximum,omitempty" json:"exclusiveMaximum,omitempty"`

	// String
	Pattern   string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	MinLength *int   `yaml:"minLength,omitempty" json:"minLength,omitempty"`
	MaxLength *int   `yaml:"maxLength,omitempty" json:"maxLength,omitempty"`
	Format    string `yaml:"format,omitempty" json:"format,omitempty"`

	// Array
	Items            *Schema   `yaml:"items,omitempty" json:"items,omitempty"`
	PrefixItems      []*Schema `yaml:"prefixItems,omitempty" json:"prefixItems,omitempty"`
	UnevaluatedItems *Schema   `yaml:"unevaluatedItems,omitempty" json:"unevaluatedItems,omitempty"`
	Contains         *Schema   `yaml:"contains,omitempty" json:"contains,omitempty"`
	MaxContains      *int      `yaml:"maxContains,omitempty" json:"maxContains,omitempty"`
	MinContains      *int      `yaml:"minContains,omitempty" json:"minContains,omitempty"`
	MinItems         *int      `yaml:"minItems,omitempty" json:"minItems,omitempty"`
	MaxItems         *int      `yaml:"maxItems,omitempty" json:"maxItems,omitempty"`
	UniqueItems      bool      `yaml:"uniqueItems,omitempty" json:"uniqueItems,omitempty"`
	ShuffleItems     bool      `yaml:"x-shuffleItems,omitempty" json:"x-shuffleItems,omitempty"`

	// Object
	Properties            *Schemas            `yaml:"properties,omitempty" json:"properties,omitempty"`
	PatternProperties     map[string]*Schema  `yaml:"patternProperties,omitempty" json:"patternProperties,omitempty"`
	MinProperties         *int                `yaml:"minProperties,omitempty" json:"minProperties,omitempty"`
	MaxProperties         *int                `yaml:"maxProperties,omitempty" json:"maxProperties,omitempty"`
	Required              []string            `yaml:"required,omitempty" json:"required,omitempty"`
	DependentRequired     map[string][]string `yaml:"dependentRequired,omitempty" json:"dependentRequired,omitempty"`
	DependentSchemas      map[string]*Schema  `yaml:"dependentSchemas,omitempty" json:"dependentSchemas,omitempty"`
	AdditionalProperties  *Schema             `yaml:"additionalProperties,omitempty" json:"additionalProperties,omitempty"`
	UnevaluatedProperties *Schema             `yaml:"unevaluatedProperties,omitempty" json:"unevaluatedProperties,omitempty"`
	PropertyNames         *Schema             `yaml:"propertyNames,omitempty" json:"propertyNames,omitempty"`

	AnyOf []*Schema `yaml:"anyOf,omitempty" json:"anyOf,omitempty"`
	AllOf []*Schema `yaml:"allOf,omitempty" json:"allOf,omitempty"`
	OneOf []*Schema `yaml:"oneOf,omitempty" json:"oneOf,omitempty"`
	Not   *Schema   `yaml:"not,omitempty" json:"not,omitempty"`

	If   *Schema `yaml:"if,omitempty" json:"if,omitempty"`
	Then *Schema `yaml:"then,omitempty" json:"then,omitempty"`
	Else *Schema `yaml:"else,omitempty" json:"else,omitempty"`

	// Annotations
	Title       string           `yaml:"title,omitempty" json:"title,omitempty"`
	Description string           `yaml:"description,omitempty" json:"description,omitempty"`
	Default     interface{}      `yaml:"default,omitempty" json:"default,omitempty"`
	Deprecated  bool             `yaml:"deprecated,omitempty" json:"deprecated,omitempty"`
	Examples    []schema.Example `yaml:"examples,omitempty" json:"examples,omitempty"`
	Example     *schema.Example  `yaml:"example,omitempty" json:"example,omitempty"`

	// Media
	ContentMediaType string `yaml:"contentMediaType,omitempty" json:"contentMediaType,omitempty"`
	ContentEncoding  string `yaml:"contentEncoding,omitempty" json:"contentEncoding,omitempty"`

	// OpenAPI
	Xml      *Xml `yaml:"xml,omitempty" json:"xml,omitempty"`
	Nullable bool `yaml:"nullable,omitempty" json:"nullable,omitempty"`

	Definitions map[string]*Schema `yaml:"definitions,omitempty" json:"definitions,omitempty"`
	Defs        map[string]*Schema `yaml:"$defs,omitempty" json:"$defs,omitempty"`

	Sub *Schema `yaml:"-" json:"-"`
	m   map[string]bool
	cm  changeManager
}

func (s *Schema) HasProperties() bool {
	return s.Properties != nil && s.Properties.Len() > 0
}

func (s *Schema) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if s == nil {
		return nil
	}

	for _, d := range s.Definitions {
		if err := d.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, d := range s.Defs {
		if err := d.Parse(config, reader); err != nil {
			return err
		}
	}

	if s.Id != "" {
		config.OpenScope(s.Id)
		defer config.CloseScope()
	} else {
		config.Scope.OpenIfNeeded(config.Info.Path())
	}

	if s.Anchor != "" {
		if err := config.Scope.SetLexical(s.Anchor, s); err != nil {
			return err
		}
	}

	if s.DynamicAnchor != "" {
		if err := config.Scope.SetDynamic(s.DynamicAnchor, s); err != nil {
			return err
		}
	}

	if err := s.Items.Parse(config, reader); err != nil {
		return err
	}

	if err := s.Properties.Parse(config, reader); err != nil {
		return err
	}

	if err := s.AdditionalProperties.Parse(config, reader); err != nil {
		return err
	}

	for _, r := range s.AnyOf {
		if err := r.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, r := range s.AllOf {
		if err := r.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, r := range s.OneOf {
		if err := r.Parse(config, reader); err != nil {
			return err
		}
	}

	if s.Ref != "" {
		err := dynamic.Resolve(s.Ref, &s.Sub, config, reader)
		if err != nil {
			return err
		}
		s.apply(s.Sub)
		s.Sub.cm.Subscribe(s.apply)
	}

	if s.DynamicRef != "" {
		err := dynamic.ResolveDynamic(s.DynamicRef, &s.Sub, config, reader)
		if err != nil {
			return err
		}
		s.apply(s.Sub)
		s.Sub.cm.Subscribe(s.apply)
	}

	s.cm.Notify(s)

	return nil
}

func (s *Schema) String() string {
	var sb strings.Builder

	if s.Boolean != nil {
		return fmt.Sprintf("%v", s.Boolean)
	}

	if len(s.AnyOf) > 0 {
		sb.WriteString("any of ")
		for _, i := range s.AnyOf {
			if sb.Len() > 7 {
				sb.WriteString(", ")
			}
			sb.WriteString(i.String())
		}
		return sb.String()
	}
	if len(s.AllOf) > 0 {
		sb.WriteString("all of ")
		for _, i := range s.AllOf {
			if sb.Len() > 7 {
				sb.WriteString(", ")
			}
			sb.WriteString(i.String())
		}
		return sb.String()
	}
	if len(s.OneOf) > 0 {
		sb.WriteString("one of ")
		for _, i := range s.OneOf {
			if sb.Len() > 7 {
				sb.WriteString(", ")
			}
			sb.WriteString(i.String())
		}
		return sb.String()
	}

	if len(s.Type) > 0 {
		sb.WriteString(fmt.Sprintf("schema type=%v", s.Type.String()))
	}

	if len(s.Format) > 0 {
		sb.WriteString(fmt.Sprintf(" format=%v", s.Format))
	}
	if len(s.Pattern) > 0 {
		sb.WriteString(fmt.Sprintf(" pattern=%v", s.Pattern))
	}
	if s.MinLength != nil {
		sb.WriteString(fmt.Sprintf(" minLength=%v", *s.MinLength))
	}
	if s.MaxLength != nil {
		sb.WriteString(fmt.Sprintf(" maxLength=%v", *s.MaxLength))
	}

	if s.ExclusiveMinimum != nil {
		if s.ExclusiveMinimum.IsA() {
			sb.WriteString(fmt.Sprintf(" exclusiveMinimum=%v", s.ExclusiveMinimum.Value()))
		} else if s.ExclusiveMinimum.B {
			sb.WriteString(fmt.Sprintf(" exclusiveMinimum=%v", *s.Minimum))
		}
	} else if s.Minimum != nil {
		sb.WriteString(fmt.Sprintf(" minimum=%v", *s.Minimum))
	}

	if s.ExclusiveMaximum != nil {
		if s.ExclusiveMaximum.IsA() {
			sb.WriteString(fmt.Sprintf(" exclusiveMaximum=%v", s.ExclusiveMaximum.Value()))
		} else if s.ExclusiveMaximum.B {
			sb.WriteString(fmt.Sprintf(" exclusiveMaximum=%v", *s.Maximum))
		}
	} else if s.Maximum != nil {
		sb.WriteString(fmt.Sprintf(" maximum=%v", *s.Maximum))
	}

	if s.MinItems != nil {
		sb.WriteString(fmt.Sprintf(" minItems=%v", *s.MinItems))
	}
	if s.MaxItems != nil {
		sb.WriteString(fmt.Sprintf(" maxItems=%v", *s.MaxItems))
	}
	if s.MinProperties != nil {
		sb.WriteString(fmt.Sprintf(" minProperties=%v", *s.MinProperties))
	}
	if s.MaxProperties != nil {
		sb.WriteString(fmt.Sprintf(" maxProperties=%v", *s.MaxProperties))
	}
	if s.UniqueItems {
		sb.WriteString(" unique-items")
	}

	if s.Type.Includes("object") && s.Properties != nil {
		var sbProp strings.Builder
		for _, p := range s.Properties.Keys() {
			if sbProp.Len() > 0 {
				sbProp.WriteString(", ")
			}
			sbProp.WriteString(fmt.Sprintf("%v", p))
		}
		sb.WriteString(fmt.Sprintf(" properties=[%v]", sbProp.String()))
	}
	if len(s.Required) > 0 {
		sb.WriteString(fmt.Sprintf(" required=%v", s.Required))
	}
	if s.Type.Includes("object") && !s.IsFreeForm() {
		sb.WriteString(" free-form=false")
	}

	if s.Type.Includes("array") && s.Items != nil {
		sb.WriteString(" items=")
		sb.WriteString(s.Items.String())
	}

	if len(s.Title) > 0 {
		sb.WriteString(fmt.Sprintf(" title=%v", s.Title))
	} else if len(s.Description) > 0 {
		sb.WriteString(fmt.Sprintf(" description=%v", s.Description))
	}

	return sb.String()
}

func (s *Schema) IsFreeForm() bool {
	if !s.Type.Includes("object") {
		return false
	}
	free := s.Type.Includes("object") && (s.Properties == nil || s.Properties.Len() == 0)
	if s.AdditionalProperties == nil || free {
		return true
	}
	if s.AdditionalProperties.Boolean != nil {
		return *s.AdditionalProperties.Boolean
	}
	return s.AdditionalProperties.IsFreeForm()
}

func (s *Schema) IsDictionary() bool {
	return s.AdditionalProperties != nil && len(s.AdditionalProperties.Type) > 0
}

func (s *Schema) IsNullable() bool {
	return s.Nullable || s.Type.IsNullable()
}

func (s *Schema) ConvertTo(i interface{}) (interface{}, error) {
	if _, ok := i.(*schema.Schema); ok {
		return ConvertToJsonSchema(s), nil
	}
	return nil, fmt.Errorf("cannot convert %v to json schema", i)
}

func (s *Schema) UnmarshalJSON(b []byte) error {
	_ = json.Unmarshal(b, &s.m)

	var boolVal bool
	if err := json.Unmarshal(b, &boolVal); err == nil {
		s.Boolean = &boolVal
		return nil
	}

	type alias Schema
	a := alias{}
	err := dynamic.UnmarshalJSON(b, &a)
	if err != nil {
		return err
	}
	a.m = s.m
	*s = Schema(a)
	return nil
}

func (s *Schema) UnmarshalYAML(node *yaml.Node) error {
	_ = node.Decode(&s.m)

	var boolVal bool
	if err := node.Decode(&boolVal); err == nil {
		s.Boolean = &boolVal
		return nil
	}

	type alias Schema
	a := alias{}
	err := node.Decode(&a)
	if err != nil {
		return err
	}
	a.m = s.m
	*s = Schema(a)
	return nil
}
