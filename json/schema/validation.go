package schema

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"mokapi/sortedmap"
	"net"
	"net/mail"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

func (s *Schema) Validate(v interface{}) error {
	if s == nil {
		return nil
	}

	if len(s.OneOf) > 0 {
		return s.validateOneOf(v)
	}
	if len(s.AllOf) > 0 {
		return s.validateAllOf(v)
	}

	if len(s.Type) == 0 {
		return nil
	}

	var errs []error
	for _, t := range s.Type {
		var err error
		switch t {
		case "boolean":
			err = s.validateBool(v)
		case "string":
			err = s.validateString(v)
		case "number", "integer":
			err = s.validateNumber(v)
		case "array":
			err = s.validateArray(v)
		case "object":
			err = s.validateObject(v)
		case "null":
			continue
		default:
			return fmt.Errorf("unsupported type %v", t)
		}
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 1 && len(s.Type) == 1 {
		return errs[0]
	}
	if len(errs) == len(s.Type) {
		return fmt.Errorf("value does not match to any type: %v", errors.Join(errs...))
	}
	return nil
}

func (r *Ref) Validate(v interface{}) error {
	if r.Value == nil {
		return nil
	}
	return r.Value.Validate(v)
}

func (s *Schema) validateBool(i interface{}) error {
	if _, ok := i.(bool); ok {
		return nil
	}
	return fmt.Errorf("validation error on %v, expected %v", i, s)
}

func (s *Schema) validateString(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("validation error on %v, expected %v", v, s)
	}

	switch s.Format {
	case "date":
		_, err := time.Parse("2006-01-02", str)
		if err != nil {
			return fmt.Errorf("value '%v' does not match format 'date' (RFC3339), expected %v", str, s)
		}
		return nil
	case "date-time":
		_, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return fmt.Errorf("value '%v' does not match format 'date-time' (RFC3339), expected %v", str, s)
		}
		return nil
	case "email":
		_, err := mail.ParseAddress(str)
		if err != nil {
			return fmt.Errorf("value '%v' does not match format 'email', expected %v", str, s)
		}
		return nil
	case "uuid":
		_, err := uuid.Parse(str)
		if err != nil {
			return fmt.Errorf("value '%v' does not match format 'uuid', expected %v", str, s)
		}
		return nil
	case "ipv4":
		ip := net.ParseIP(str)
		if ip == nil {
			return fmt.Errorf("value '%v' does not match format 'ipv4', expected %v", str, s)
		}
		if len(strings.Split(str, ".")) != 4 {
			return fmt.Errorf("value '%v' does not match format 'ipv4', expected %v", str, s)
		}
		return nil
	case "ipv6":
		ip := net.ParseIP(str)
		if ip == nil {
			return fmt.Errorf("value '%v' does not match format 'ipv6', expected %v", str, s)
		}
		if len(strings.Split(str, ":")) != 8 {
			return fmt.Errorf("value '%v' does not match format 'ipv6', expected %v", str, s)
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
		return fmt.Errorf("value '%v' does not match pattern, expected %v", str, s)
	}

	if s.MinLength != nil && *s.MinLength > len(str) {
		return fmt.Errorf("length of '%v' is too short, expected %v", str, s)
	}
	if s.MaxLength != nil && *s.MaxLength < len(str) {
		return fmt.Errorf("length of '%v' is too long, expected %v", str, s)
	}

	if len(s.Enum) > 0 {
		return checkValueIsInEnum(str, s.Enum)
	}

	return nil
}

func (s *Schema) validateNumber(v interface{}) error {
	var n float64

	switch i := v.(type) {
	case int:
		n = float64(i)
	case int32:
		n = float64(i)
	case int64:
		n = float64(i)
	case float32:
		n = float64(i)
	case float64:
		n = i
	default:
		return fmt.Errorf("validation error on %v, expected %v", v, s)
	}

	switch {
	case s.IsNumber() && s.Format == "float":
		if float64(float32(n)) != n {
			return fmt.Errorf("validation error on %v, expected %v", v, s)
		}
	case s.IsInteger() && s.Format == "int32":
		if float64(int32(n)) != n {
			return fmt.Errorf("validation error on %v, expected %v", v, s)
		}
	case s.IsInteger():
		if float64(int64(n)) != n {
			return fmt.Errorf("validation error on %v, expected %v", v, s)
		}
	}

	if s.Minimum != nil && n < *s.Minimum {
		return fmt.Errorf("%v is lower as required minimum %v, expected %v", n, *s.Minimum, s)
	}
	if s.ExclusiveMinimum != nil && n <= *s.ExclusiveMinimum {
		return fmt.Errorf("%v is lower or equal as required minimum %v, expected %v", n, *s.ExclusiveMinimum, s)
	}

	if s.Maximum != nil && n < *s.Maximum {
		return fmt.Errorf("%v is greater as required maximum %v, expected %v", n, *s.Maximum, s)
	}
	if s.ExclusiveMaximum != nil && n <= *s.ExclusiveMaximum {
		return fmt.Errorf("%v is greater or equal as the required maximum %v, expected %v", n, *s.ExclusiveMaximum, s)
	}

	if len(s.Enum) > 0 {
		return checkValueIsInEnum(n, s.Enum)
	}

	if s.MultipleOf != nil {
		m := n / float64(*s.MultipleOf)
		if m != float64(int(m)) {
			return fmt.Errorf("%v is not a multiple of %v: %v", n, *s.MultipleOf, s)
		}
	}

	return nil
}

func (s *Schema) validateArray(i interface{}) error {
	v := reflect.ValueOf(i)
	if s.MinItems != nil && v.Len() < *s.MinItems {
		return fmt.Errorf("should NOT have less than %v items", *s.MinItems)
	}
	if s.MaxItems != nil && v.Len() > *s.MaxItems {
		return fmt.Errorf("should NOT have more than %v items", *s.MaxItems)
	}

	if len(s.Enum) > 0 {
		return checkValueIsInEnum(i, s.Enum)
	}

	if s.UniqueItems {
		var unique []interface{}
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			for _, u := range unique {
				if reflect.DeepEqual(item, u) {
					return fmt.Errorf("should NOT have duplicate items (%v)", toString(item))
				}
			}
			unique = append(unique, item)
		}
	}

	return nil
}

func (s *Schema) validateObject(v interface{}) error {
	m, ok := v.(map[string]interface{})
	if !ok {
		lm, ok := v.(*sortedmap.LinkedHashMap[string, interface{}])
		if !ok {
			return fmt.Errorf("expected object but got %v", reflect.TypeOf(v))
		}
		m = lm.ToMap()
	}

	if s.MinProperties != nil && len(m) < *s.MinProperties {
		return fmt.Errorf("number of properties (%v) is lower than minimum %v: %v", len(m), *s.MinProperties, toString(v))
	}
	if s.MaxProperties != nil && len(m) > *s.MaxProperties {
		return fmt.Errorf("number of properties (%v) is greater than maximum %v: %v", len(m), *s.MaxProperties, toString(v))
	}
	if s.AdditionalProperties.Forbidden {
		var additionals []string
		for name := range m {
			if _, ok := s.Properties.Get(name); !ok {
				additionals = append(additionals, name)
			}
		}
		if len(additionals) > 0 {
			sort.Strings(additionals)
			return fmt.Errorf("additional properties not allowed: %v, expected %v", strings.Join(additionals, ", "), s)
		}
	}
	var missing []string
	for _, p := range s.Required {
		if _, ok := m[p]; !ok {
			missing = append(missing, p)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required field(s) %v", missing)
	}

	if len(s.Enum) > 0 {
		return checkValueIsInEnum(v, s.Enum)
	}

	for k, v := range m {
		if prop, ok := s.Properties.Get(k); ok && prop.Value != nil {
			if err := prop.Value.Validate(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Schema) validateOneOf(v interface{}) error {
	var result interface{}

	for _, ref := range s.OneOf {
		// free-form object
		if ref.IsObject() && ref.Value.Properties == nil {
			result = v
			continue
		}

		err := ref.Validate(v)
		if err != nil {
			continue
		}

		if result != nil {
			return fmt.Errorf("validation error %v: it is valid for more than one schema, expected %v", toString(v), s)
		}

		result = v
	}

	if result == nil {
		return fmt.Errorf("validation error %v: expected to match one of schema but it matches none", toString(v))
	}

	return nil
}

func (s *Schema) validateAllOf(v interface{}) error {
	for _, r := range s.AllOf {
		one := r.Value

		// free-form object
		if one.Properties == nil {
			return nil
		}

		err := one.Validate(v)
		if err != nil {
			return fmt.Errorf("validation error %v: value does not match part of allOf: %w", toString(v), err)
		}
	}

	return nil
}

func checkValueIsInEnum(v interface{}, enum []interface{}) error {
	found := false
	for _, x := range enum {
		if reflect.DeepEqual(v, x) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("validation error %v: value does not match one in the enumeration %v", toString(v), toString(enum))
	}

	return nil
}

func toString(i interface{}) string {
	var sb strings.Builder
	switch o := i.(type) {
	case []interface{}:
		sb.WriteRune('[')
		for i, v := range o {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(toString(v))
		}
		sb.WriteRune(']')
	case map[string]interface{}:
		sb.WriteRune('{')
		for key, val := range o {
			if sb.Len() > 1 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%v: %v", key, toString(val)))
		}
		sb.WriteRune('}')
	case string, int, int32, int64, float32, float64:
		sb.WriteString(fmt.Sprintf("%v", o))
	case *sortedmap.LinkedHashMap[string, interface{}]:
		return o.String()
	default:
		v := reflect.ValueOf(i)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		t := reflect.TypeOf(i)
		switch v.Kind() {
		case reflect.Slice:
			sb.WriteRune('[')
			for i := 0; i < v.Len(); i++ {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(toString(v.Index(i).Interface()))
			}
			sb.WriteRune(']')
		case reflect.Struct:
			sb.WriteRune('{')
			for i := 0; i < v.NumField(); i++ {
				if i > 0 {
					sb.WriteString(", ")
				}
				name := t.Field(i).Name
				fv := v.Field(i).Interface()
				sb.WriteString(fmt.Sprintf("%v: %v", name, fv))
			}
			sb.WriteRune('}')
		}
	}
	return sb.String()
}
