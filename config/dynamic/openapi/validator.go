package openapi

import (
	"fmt"
	"github.com/google/uuid"
	"math"
	"net"
	"net/mail"
	"reflect"
	"regexp"
	"strings"
	"time"
)

func Validate(i interface{}, s *Schema) error {
	if s == nil || len(s.Type) == 0 {
		return nil
	}

	if s.Enum != nil {
		for _, e := range s.Enum {
			if reflect.DeepEqual(i, e) {
				return nil
			}
		}
		return fmt.Errorf("value does not match any enum value")
	}

	switch s.Type {
	case "string":
		return validateString(i, s)
	case "integer":
		return validateInteger(i, s)
	case "number":
		return validateNumber(i, s)
	case "object":
		return validateObject(i, s)
	case "array":
		return validateArray(i, s)
	}

	return fmt.Errorf("unsupported type %v", s.Type)
}

func validateObject(i interface{}, s *Schema) error {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected object got %v", v.Kind())
	}

	required := make(map[string]struct{})
	for _, r := range s.Required {
		required[r] = struct{}{}
	}

	// free-form object
	if s.Properties == nil {
		return nil
	}

	if v.NumField() > s.Properties.Value.Len() {
		return fmt.Errorf("too many properties for object")
	}

	for it := s.Properties.Value.Iter(); it.Next(); {
		name := it.Key().(string)
		p := it.Value().(*SchemaRef).Value
		f := v.FieldByName(strings.Title(name))
		if !f.IsValid() || f.IsZero() {
			if _, ok := required[name]; ok && len(required) > 0 {
				return fmt.Errorf("expected required property %v", name)
			}
			continue
		}
		err := Validate(f.Interface(), p)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateArray(i interface{}, s *Schema) error {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("expected array got %v", v.Kind())
	}

	a, ok := i.([]interface{})
	if !ok {
		return fmt.Errorf("expected []inteface{} for type array")
	}

	if s.MinItems != nil && *s.MinItems > len(a) {
		return fmt.Errorf("array requires minimum items of %v", *s.MinItems)
	}
	if s.MaxItems != nil && *s.MaxItems < len(a) {
		return fmt.Errorf("array requires maximum items of %v", *s.Maximum)
	}

	if !s.UniqueItems && s.Items == nil {
		return nil
	}

	m := make(map[interface{}]struct{})
	for _, item := range a {
		if _, ok := m[item]; ok && s.UniqueItems {
			return fmt.Errorf("array requires unique items")
		}
		m[item] = struct{}{}

		if s.Items != nil {
			err := Validate(item, s.Items.Value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func validateString(i interface{}, s *Schema) error {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.String {
		return fmt.Errorf("expected string got %v", v.Kind())
	}

	str := i.(string)
	switch s.Format {
	case "date":
		_, err := time.Parse("2006-01-02", str)
		if err != nil {
			return fmt.Errorf("string is not a date RFC3339")
		}
		return nil
	case "date-time":
		_, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return fmt.Errorf("string is not a date-time RFC3339")
		}
		return nil
	case "email":
		_, err := mail.ParseAddress(str)
		if err != nil {
			return fmt.Errorf("string is not an email address")
		}
		return nil
	case "uuid":
		_, err := uuid.Parse(str)
		if err != nil {
			return fmt.Errorf("string is not an uuid")
		}
		return nil
	case "ipv4":
		ip := net.ParseIP(str)
		if ip == nil {
			return fmt.Errorf("string is not an ipv4")
		}
		if len(strings.Split(str, ".")) != 4 {
			return fmt.Errorf("string is not an ipv4")
		}
		return nil
	case "ipv6":
		ip := net.ParseIP(str)
		if ip == nil {
			return fmt.Errorf("string is not an ipv6")
		}
		if len(strings.Split(str, ":")) != 8 {
			return fmt.Errorf("string is not an ipv6")
		}
		return nil
	}

	if len(s.Pattern) > 0 {
		p, err := regexp.Compile(s.Pattern)
		if err != nil {
			return err
		}
		if p.MatchString(str) {
			return nil
		}
		return fmt.Errorf("value does not match pattern")
	}

	return nil
}

func validateInteger(i interface{}, s *Schema) error {
	v := reflect.ValueOf(i)
	var n int64
	switch v.Kind() {
	case reflect.Int:
		n = int64(i.(int))
	case reflect.Int32:
		n = int64(i.(int32))
	case reflect.Int64:
		n = i.(int64)
	default:
		return fmt.Errorf("expected integer got %v", v.Kind())
	}

	switch s.Format {
	case "int32":
		if n > math.MaxInt32 || n < math.MinInt32 {
			return fmt.Errorf("integer is not int32")
		}
	}

	if s.Minimum != nil && n < int64(*s.Minimum) {
		return fmt.Errorf("value is lower as defined minimum %v", *s.Minimum)
	}

	if s.Maximum != nil && n > int64(*s.Maximum) {
		return fmt.Errorf("value is greater as defined maximum %v", *s.Maximum)
	}
	return nil
}

func validateNumber(i interface{}, s *Schema) error {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
	default:
		return fmt.Errorf("expected number got %v", v.Kind())
	}

	n := i.(float64)
	switch s.Format {
	case "float32":
		if n > math.MaxFloat32 || n < -math.MaxFloat32 {
			return fmt.Errorf("number is not float32")
		}
	}

	if s.Minimum != nil && n < *s.Minimum {
		return fmt.Errorf("value is lower as defined minimum %v", *s.Minimum)
	}

	if s.Maximum != nil && n > *s.Maximum {
		return fmt.Errorf("value is greater as defined maximum %v", *s.Maximum)
	}
	return nil
}
