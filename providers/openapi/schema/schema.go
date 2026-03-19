package schema

import (
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/schema/json/schema"

	"gopkg.in/yaml.v3"
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
	UniqueItems      *bool     `yaml:"uniqueItems,omitempty" json:"uniqueItems,omitempty"`
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
}

func (s *Schema) HasProperties() bool {
	return s != nil && s.Properties != nil && s.Properties.Len() > 0
}

func (s *Schema) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if s == nil {
		return nil
	}

	if s.Id != "" {
		config.OpenScope(s.Id)
		defer config.CloseScope()
	} else {
		config.Scope.OpenIfNeeded(config.Info.Path())
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

	if !s.skipParse("items") {
		if err := s.Items.Parse(config, reader); err != nil {
			return err
		}
	}

	if s.Properties != nil && !s.skipParse("properties") {
		for it := s.Properties.Iter(); it.Next(); {
			if err := it.Value().Parse(config, reader); err != nil {
				return fmt.Errorf("parse schema '%v' failed: %w", it.Key(), err)
			}
		}
	}

	if !s.skipParse("additionalProperties") {
		if err := s.AdditionalProperties.Parse(config, reader); err != nil {
			return err
		}
	}

	if !s.skipParse("anyOf") {
		for _, r := range s.AnyOf {
			if err := r.Parse(config, reader); err != nil {
				return err
			}
		}
	}

	if !s.skipParse("allOf") {
		for _, r := range s.AllOf {
			if err := r.Parse(config, reader); err != nil {
				return err
			}
		}
	}

	if !s.skipParse("oneOf") {
		for _, r := range s.OneOf {
			if err := r.Parse(config, reader); err != nil {
				return err
			}
		}
	}

	if s.Ref != "" {
		err := dynamic.Resolve(s.Ref, &s.Sub, config, reader)
		if err != nil {
			return err
		}

		// Apply the resolved schema as an overlay onto the current schema.
		// The referenced schema is cloned to preserve the immutability of
		// the parsed schema graph. Dynamic references may resolve differently
		// depending on the evaluation context, so shared schema nodes must
		// never be mutated.
		s.apply(s.Sub)
	}

	if s.DynamicRef != "" {
		err := dynamic.ResolveDynamic(s.DynamicRef, &s.Sub, config, reader)
		if err != nil {
			return err
		}
		s.apply(s.Sub)
	}

	return nil
}

func (s *Schema) String() string {
	return ConvertToJsonSchema(s).String()
}

func (s *Schema) IsFreeForm() bool {
	if s == nil {
		return true
	}
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
	m := map[string]json.RawMessage{}
	_ = json.Unmarshal(b, &m)
	if s.m == nil {
		s.m = map[string]bool{}
	}
	for k := range m {
		s.m[k] = true
	}

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
	m := map[string]yaml.Node{}
	_ = node.Decode(&m)
	if s.m == nil {
		s.m = map[string]bool{}
	}
	for k := range m {
		s.m[k] = true
	}

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

func (s *Schema) skipParse(name string) bool {
	if s.Ref != "" {
		_, ok := s.m[name]
		return !ok
	}
	return false
}
