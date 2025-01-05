package schema

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

type Types []string

//goland:noinspection GoMixedReceiverTypes
func (t Types) String() string {
	if t == nil || len(t) == 0 {
		return ""
	}
	if len(t) == 1 {
		return (t)[0]
	}

	var sb strings.Builder
	sb.WriteString("[")
	for i, s := range t {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(s)
	}
	sb.WriteString("]")
	return sb.String()
}

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

func (t *Types) IsBool() bool {
	return t.Includes("boolean")
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

func (s *Schema) HasProperties() bool {
	if s == nil {
		return false
	}
	return s.Properties != nil && s.Properties.Len() > 0
}

func (s *Schema) Is(typeName string) bool {
	return s != nil && s.Type.Includes(typeName)
}

func (s *Schema) IsFreeForm() bool {
	if s == nil {
		return false
	}
	if !s.IsObject() && len(s.Type) > 0 {
		return false
	}
	if s.AdditionalProperties == nil {
		return true
	}
	if s.AdditionalProperties.Boolean != nil {
		return *s.AdditionalProperties.Boolean
	}

	return s.AdditionalProperties.Value == nil
}

func (s *Schema) IsAnyString() bool {
	if !s.IsString() {
		return false
	}
	return s.Pattern == "" && s.Format == "" && s.MinLength == nil && s.MaxLength == nil
}

func (s *Schema) IsFalse() bool {
	return s != nil && s.Boolean != nil && !*s.Boolean
}

func (t *Types) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte(""), nil
	}

	if len(*t) == 1 {
		return []byte(fmt.Sprintf(`"%v"`, (*t)[0])), nil
	}
	var sb strings.Builder
	for _, item := range *t {
		if sb.Len() > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf(`"%v"`, item))
	}
	return []byte(fmt.Sprintf("[%v]", sb.String())), nil
}
