package schema

import (
	"encoding/json"
	"fmt"
)

type Schema struct {
	Type  Types         `yaml:"type,omitempty" json:"type,omitempty"`
	Enum  []interface{} `yaml:"enum,omitempty" json:"enum,omitempty"`
	Const interface{}   `yaml:"const,omitempty" json:"const,omitempty"`

	// Numbers
	MultipleOf       *float64 `yaml:"multipleOf,omitempty" json:"multipleOf,omitempty"`
	Maximum          *float64 `yaml:"maximum,omitempty" json:"maximum,omitempty"`
	ExclusiveMaximum *float64 `yaml:"exclusiveMaximum,omitempty" json:"exclusiveMaximum,omitempty"`
	Minimum          *float64 `yaml:"minimum,omitempty" json:"minimum,omitempty"`
	ExclusiveMinimum *float64 `yaml:"exclusiveMinimum,omitempty" json:"ExclusiveMinimum,omitempty"`

	// Strings
	MaxLength *int   `yaml:"maxLength,omitempty" json:"maxLength,omitempty"`
	MinLength *int   `yaml:"minLength,omitempty" json:"minLength,omitempty"`
	Pattern   string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Format    string `yaml:"format,omitempty" json:"format,omitempty"`

	// Arrays
	Items        *Ref `yaml:"items,omitempty" json:"items,omitempty"`
	MaxItems     *int `yaml:"maxItems,omitempty" json:"maxItems,omitempty"`
	MinItems     *int `yaml:"minItems,omitempty" json:"minItems,omitempty"`
	UniqueItems  bool `yaml:"uniqueItems,omitempty" json:"uniqueItems,omitempty"`
	MaxContains  *int `yaml:"maxContains,omitempty" json:"maxContains,omitempty"`
	MinContains  *int `yaml:"minContains,omitempty" json:"minContains,omitempty"`
	ShuffleItems bool `yaml:"x-shuffleItems,omitempty" json:"x-shuffleItems,omitempty"`

	// Objects
	Properties           *Schemas             `yaml:"properties,omitempty" json:"properties,omitempty"`
	MaxProperties        *int                 `yaml:"maxProperties,omitempty" json:"maxProperties,omitempty"`
	MinProperties        *int                 `yaml:"minProperties,omitempty" json:"minProperties,omitempty"`
	Required             []string             `yaml:"required,omitempty" json:"required,omitempty"`
	DependentRequired    map[string][]string  `yaml:"dependentRequired,omitempty" json:"dependentRequired,omitempty"`
	AdditionalProperties AdditionalProperties `yaml:"additionalProperties,omitempty" json:"additionalProperties,omitempty"`

	AllOf []*Ref `yaml:"allOf,omitempty" json:"allOf,omitempty"`
	AnyOf []*Ref `yaml:"anyOf,omitempty" json:"anyOf,omitempty"`
	OneOf []*Ref `yaml:"oneOf,omitempty" json:"oneOf,omitempty"`
}

type UnmarshalError struct {
	Value interface{}
	Field string
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("cannot unmarshal %v into field %v of type schema", e.Value, e.Field)
}

func (s *Schema) Parse() error {
	if s.MultipleOf != nil && *s.MultipleOf <= 0 {
		return fmt.Errorf("multipleOf must be greater than 0: %v", *s.MultipleOf)
	}
	if s.MaxLength != nil && *s.MaxLength < 0 {
		return fmt.Errorf("maxLength must be a non-negative integer: %v", *s.MaxLength)
	}
	if s.MinLength != nil && *s.MinLength < 0 {
		return fmt.Errorf("minLength must be a non-negative integer: %v", *s.MinLength)
	}
	if s.MinLength != nil && s.MaxLength != nil && *s.MinLength > *s.MaxLength {
		return fmt.Errorf("minLength cannot be greater than maxLength: %v, %v", *s.MinLength, *s.MaxLength)
	}
	if s.MaxItems != nil && *s.MaxItems < 0 {
		return fmt.Errorf("maxItems must be a non-negative integer: %v", *s.MaxItems)
	}
	if s.MinItems != nil && *s.MinItems < 0 {
		return fmt.Errorf("minItems must be a non-negative integer: %v", *s.MinItems)
	}
	if s.MinItems != nil && s.MaxItems != nil && *s.MinItems > *s.MaxItems {
		return fmt.Errorf("minItems cannot be greater than maxItems: %v, %v", *s.MinItems, *s.MaxItems)
	}
	if s.MaxContains != nil && *s.MaxContains < 0 {
		return fmt.Errorf("maxContains must be a non-negative integer: %v", *s.MaxContains)
	}
	if s.MinContains != nil && *s.MinContains < 0 {
		return fmt.Errorf("minContains must be a non-negative integer: %v", *s.MinContains)
	}
	if s.MaxProperties != nil && *s.MaxProperties < 0 {
		return fmt.Errorf("maxProperties must be a non-negative integer: %v", *s.MaxProperties)
	}
	if s.MinProperties != nil && *s.MinProperties < 0 {
		return fmt.Errorf("minProperties must be a non-negative integer: %v", *s.MinProperties)
	}
	if s.MinProperties != nil && s.MaxProperties != nil && *s.MinProperties > *s.MaxProperties {
		return fmt.Errorf("minProperties cannot be greater than maxProperties: %v, %v", *s.MinProperties, *s.MaxProperties)
	}
	return nil
}

func (s *Schema) UnmarshalJSON(b []byte) error {
	type alias Schema
	a := alias{}
	err := json.Unmarshal(b, &a)
	if typeErr, ok := err.(*json.UnmarshalTypeError); ok {
		return &UnmarshalError{
			Value: typeErr.Value,
			Field: typeErr.Field,
		}
	} else if err != nil {
		return err
	}
	*s = Schema(a)
	return nil

	/*var raw map[string]float64
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}
	/*for k, v := range raw {
		switch k {
		case "type":
			if str, ok := v.(string); ok {
				s.Type = append(s.Type, str)
			} else if arr, ok := v.([]interface{}); ok {
				for _, i := range arr {
					if str, ok := i.(string); ok {
						s.Type = append(s.Type, str)
					} else {
						return &UnmarshalError{Value: i, Field: "type"}
					}
				}
			} else {
				return &UnmarshalError{Value: v, Field: "type"}
			}

		case "enum":
			if enum, ok := v.([]interface{}); ok {
				s.Enum = enum
			} else {
				return &UnmarshalError{Value: v, Field: "enum"}
			}
		case "const":
			s.Const = v
		case "multipleOf":
			multipleOf, err := toInt(v)
			if err != nil {
				return err
			}
			s.MultipleOf = &multipleOf
		case "maximum":
			if f, ok := v.(float64); ok {
				s.Maximum = &f
			} else {
				return &UnmarshalError{Value: v, Field: "maximum"}
			}
		case "exclusiveMaximum":
			if f, ok := v.(float64); ok {
				s.ExclusiveMaximum = &f
			} else {
				return &UnmarshalError{Value: v, Field: "exclusiveMaximum"}
			}
		case "minimum":
			if f, ok := v.(float64); ok {
				s.Minimum = &f
			} else {
				return &UnmarshalError{Value: v, Field: "minimum"}
			}
		case "exclusiveMinimum":
			if f, ok := v.(float64); ok {
				s.ExclusiveMinimum = &f
			} else {
				return &UnmarshalError{Value: v, Field: "exclusiveMinimum"}
			}
		case "maxLength":
			maxLength, err := toInt(v)
			if err != nil {
				return err
			}
			if maxLength < 0 {
				return fmt.Errorf("maxLength must be a non-negative integer: %v", maxLength)
			}
			s.MaxLength = &maxLength
		case "minLength":
			minLength, err := toInt(v)
			if err != nil {
				return err
			}
			if minLength < 0 {
				return fmt.Errorf("minLength must be a non-negative integer: %v", minLength)
			}
			s.MinLength = &minLength
		case "pattern":
			if str, ok := v.(string); ok {
				s.Pattern = str
			} else {
				return &UnmarshalError{Value: v, Field: "pattern"}
			}
		case "format":
			if str, ok := v.(string); ok {
				s.Format = str
			} else {
				return &UnmarshalError{Value: v, Field: "format"}
			}
		case "maxItems":
			maxItems, err := toInt(v)
			if err != nil {
				return err
			}
			if maxItems < 0 {
				return fmt.Errorf("maxItems must be a non-negative integer: %v", maxItems)
			}
			s.MaxItems = &maxItems
		case "minItems":
			minItems, err := toInt(v)
			if err != nil {
				return err
			}
			if minItems < 0 {
				return fmt.Errorf("minItems must be a non-negative integer: %v", minItems)
			}
			s.MinItems = &minItems
		case "uniqueItems":
			if b, ok := v.(bool); ok {
				s.UniqueItems = b
			} else {
				return &UnmarshalError{Value: v, Field: "uniqueItems"}
			}
		case "maxContains":
			s.MaxContains, err = toInt(v)
			if err != nil {
				return err
			}
			if s.MaxContains < 0 {
				return fmt.Errorf("maxContains must be a non-negative integer: %v", s.MaxContains)
			}
		case "minContains":
			s.MinContains, err = toInt(v)
			if err != nil {
				return err
			}
			if s.MinContains < 0 {
				return fmt.Errorf("minContains must be a non-negative integer: %v", s.MinContains)
			}
		case "properties":
			m, ok := v.(map[string]interface{})
			_ = m
			_ = ok
		case "maxProperties":
			maxProperties, err := toInt(v)
			if err != nil {
				return err
			}
			if maxProperties < 0 {
				return fmt.Errorf("maxProperties must be a non-negative integer: %v", maxProperties)
			}
			s.MaxProperties = &maxProperties
		case "minProperties":
			minProperties, err := toInt(v)
			if err != nil {
				return err
			}
			if minProperties < 0 {
				return fmt.Errorf("minProperties must be a non-negative integer: %v", minProperties)
			}
			s.MinProperties = &minProperties
		case "required":
			if arr, ok := v.([]interface{}); ok {
				for _, i := range arr {
					if str, ok := i.(string); ok {
						s.Required = append(s.Required, str)
					} else {
						return &UnmarshalError{Value: i, Field: "required"}
					}
				}
			} else {
				return &UnmarshalError{Value: v, Field: "required"}
			}
		case "dependentRequired":
			s.DependentRequired = map[string][]string{}
			if m, ok := v.(map[string]interface{}); ok {
				for k, list := range m {
					if props, ok := list.([]interface{}); ok {
						for _, prop := range props {
							if str, ok := prop.(string); ok {
								s.DependentRequired[k] = append(s.DependentRequired[k], str)
							} else {
								return &UnmarshalError{Value: v, Field: "dependentRequired"}
							}
						}
					} else {
						return &UnmarshalError{Value: v, Field: "dependentRequired"}
					}
				}
			} else {
				return &UnmarshalError{Value: v, Field: "dependentRequired"}
			}
		}

	}*/
	return nil
}

func toInt(v interface{}) (int, error) {
	if f, ok := v.(float64); ok {
		i := int(f)
		if float64(i) != f {
			return 0, &UnmarshalError{Value: v, Field: "multipleOf"}
		}
		return i, nil
	} else {
		return 0, &UnmarshalError{Value: v, Field: "multipleOf"}
	}
}
