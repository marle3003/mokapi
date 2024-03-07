package schema

import (
	"encoding/json"
	"fmt"
)

type Schema struct {
	Type  []string
	Enum  []interface{}
	Const interface{}

	// Numbers
	MultipleOf       *int
	Maximum          *float64
	ExclusiveMaximum *float64
	Minimum          *float64
	ExclusiveMinimum *float64

	// Strings
	MaxLength *int
	MinLength *int
	Pattern   string
	Format    string

	// Arrays
	Items        *Ref
	MaxItems     *int
	MinItems     *int
	UniqueItems  bool
	MaxContains  int
	MinContains  int
	ShuffleItems bool

	// Objects
	Properties           *Schemas
	MaxProperties        *int
	MinProperties        *int
	Required             []string
	DependentRequired    map[string][]string
	AdditionalProperties AdditionalProperties

	AllOf []*Ref
	AnyOf []*Ref
	OneOf []*Ref
}

type UnmarshalError struct {
	Value interface{}
	Field string
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("cannot unmarshal %v into field %v of type schema", e.Value, e.Field)
}

func (s *Schema) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}
	for k, v := range raw {
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

	}
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
