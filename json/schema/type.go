package schema

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

type Types []string

func (t *Types) UnmarshalJSON(b []byte) error {
	var v interface{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	if str, ok := v.(string); ok {
		*t = append(*t, str)
	} else if arr, ok := v.([]interface{}); ok {
		for _, i := range arr {
			if str, ok := i.(string); ok {
				*t = append(*t, str)
			} else {
				return &UnmarshalError{Value: i, Field: "type"}
			}
		}
	} else {
		return &UnmarshalError{Value: v, Field: "type"}
	}
	return nil
}

func (t *Types) UnmarshalYAML(value *yaml.Node) error {
	var v interface{}
	err := value.Decode(&v)
	if err != nil {
		return err
	}
	if str, ok := v.(string); ok {
		*t = append(*t, str)
	} else if arr, ok := v.([]interface{}); ok {
		for _, i := range arr {
			if str, ok := i.(string); ok {
				*t = append(*t, str)
			} else {
				return &UnmarshalError{Value: i, Field: "type"}
			}
		}
	} else {
		return &UnmarshalError{Value: v, Field: "type"}
	}
	return nil
}

func (t *Types) Includes(typeName string) bool {
	if t == nil {
		return false
	}
	for _, v := range *t {
		if v == typeName {
			return true
		}
	}
	return false
}

func (t *Types) IsOneOf(typeNames ...string) bool {
	for _, typeName := range typeNames {
		if t.Includes(typeName) {
			return true
		}
	}
	return false
}

func (t *Types) IsAny() bool {
	return len(*t) == 0
}

func (t *Types) IsString() bool {
	return t.Includes("string")
}

func (t *Types) IsInteger() bool {
	return t.Includes("integer")
}

func (t *Types) IsNumber() bool {
	return t.Includes("number")
}

func (t *Types) IsArray() bool {
	return t.Includes("array")
}

func (t *Types) IsObject() bool {
	return t.Includes("object")
}

func (t *Types) IsNullable() bool {
	return t.Includes("null")
}

func (s *Schema) IsAny() bool {
	return s == nil || s.Type.IsAny()
}

func (s *Schema) IsString() bool {
	return s.Is("string")
}

func (s *Schema) IsInteger() bool {
	return s.Is("integer")
}

func (s *Schema) IsNumber() bool {
	return s.Is("number")
}

func (s *Schema) IsArray() bool {
	return s.Is("array")
}

func (s *Schema) IsObject() bool {
	return s.Is("object")
}

func (s *Schema) IsNullable() bool {
	return s.Is("null")
}

func (s *Schema) IsDictionary() bool {
	if s == nil {
		return false
	}
	return s.AdditionalProperties.Ref != nil && s.AdditionalProperties.Value != nil
}

func (s *Schema) HasProperties() bool {
	if s == nil {
		return false
	}
	return s.Properties != nil && s.Properties.Len() > 0
}

func (s *Schema) Is(typeName string) bool {
	return s != nil && s.Type.Includes(typeName)
}

func (s *Schema) IsFreeFrom() bool {
	return s.IsObject() && s.Properties == nil || s.Properties.Len() == 0
}

func (s *Schema) IsAnyString() bool {
	if !s.IsString() {
		return false
	}
	return s.Pattern == "" && s.Format == "" && s.MinLength == nil && s.MaxLength == nil
}
