package schema

import (
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"mokapi/sortedmap"
	"net"
	"net/mail"
	"reflect"
	"regexp"
	"strings"
	"time"
)

func validateString(i interface{}, s *Schema) error {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.String {
		return fmt.Errorf("validation error on %v, expected %v", v.Kind(), s)
	}

	str := i.(string)
	switch s.Format {
	case "date":
		_, err := time.Parse("2006-01-02", str)
		if err != nil {
			return fmt.Errorf("value '%v' is not a date RFC3339, expected %v", str, s)
		}
		return nil
	case "date-time":
		_, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return fmt.Errorf("value '%v' is not a date-time RFC3339, expected %v", str, s)
		}
		return nil
	case "email":
		_, err := mail.ParseAddress(str)
		if err != nil {
			return fmt.Errorf("value '%v' is not an email address, expected %v", str, s)
		}
		return nil
	case "uuid":
		_, err := uuid.Parse(str)
		if err != nil {
			return fmt.Errorf("value '%v' is not an uuid, expected %v", str, s)
		}
		return nil
	case "ipv4":
		ip := net.ParseIP(str)
		if ip == nil {
			return fmt.Errorf("value '%v' is not an ipv4, expected %v", str, s)
		}
		if len(strings.Split(str, ".")) != 4 {
			return fmt.Errorf("value '%v' is not an ipv4, expected %v", str, s)
		}
		return nil
	case "ipv6":
		ip := net.ParseIP(str)
		if ip == nil {
			return fmt.Errorf("value '%v' is not an ipv6, expected %v", str, s)
		}
		if len(strings.Split(str, ":")) != 8 {
			return fmt.Errorf("value '%v' is not an ipv6, expected %v", str, s)
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

	if s.MinLength > len(str) {
		return fmt.Errorf("value '%v' does not meet min length of %v", str, s.MinLength)
	}
	if s.MaxLength != nil && *s.MaxLength < len(str) {
		return fmt.Errorf("value '%v' does not meet max length of %v", str, *s.MaxLength)
	}

	if len(s.Enum) > 0 {
		return checkValueIsInEnum(str, s.Enum, &Schema{Type: "string"})
	}

	return nil
}

func validateFloat64(n float64, schema *Schema) error {
	if schema.Minimum != nil {
		min := *schema.Minimum
		if schema.ExclusiveMinimum != nil && (*schema.ExclusiveMinimum) && n <= min {
			return fmt.Errorf("%v is lower or equal as the required minimum %v, expected %v", n, min, schema)
		} else if n < min {
			return fmt.Errorf("%v is lower as the required minimum %v, expected %v", n, min, schema)
		}
	}
	if schema.Maximum != nil {
		max := *schema.Maximum
		if schema.ExclusiveMaximum != nil && (*schema.ExclusiveMaximum) && n >= max {
			return fmt.Errorf("%v is greater or equal as the required maximum %v, expected %v", n, max, schema)
		} else if n > max {
			return fmt.Errorf("%v is greater as the required maximum %v, expected %v", n, max, schema)
		}
	}

	if len(schema.Enum) > 0 {
		return checkValueIsInEnum(n, schema.Enum, &Schema{Type: "number", Format: "double"})
	}

	return nil
}

func validateInt64(n int64, schema *Schema) error {
	if schema.Minimum != nil {
		min := int64(*schema.Minimum)
		if schema.ExclusiveMinimum != nil && (*schema.ExclusiveMinimum) && n <= min {
			return fmt.Errorf("%v is lower or equal as the required minimum %v, expected %v", n, min, schema)
		} else if n < min {
			return fmt.Errorf("%v is lower as the required minimum %v, expected %v", n, min, schema)
		}
	}
	if schema.Maximum != nil {
		max := int64(*schema.Maximum)
		if schema.ExclusiveMaximum != nil && (*schema.ExclusiveMaximum) && n >= max {
			return fmt.Errorf("%v is greater or equal as the required maximum %v, expected %v", n, max, schema)
		} else if n > max {
			return fmt.Errorf("%v is greater as the required maximum %v, expected %v", n, max, schema)
		}
	}

	if len(schema.Enum) > 0 {
		return checkValueIsInEnum(n, schema.Enum, &Schema{Type: "integer", Format: "int64"})
	}

	return nil
}

func validateArray(a interface{}, schema *Schema) error {
	v := reflect.ValueOf(a)
	if schema.MinItems != nil && v.Len() < *schema.MinItems {
		return fmt.Errorf("validation error minItems on %v, expected %v", toString(a), schema)
	}
	if schema.MaxItems != nil && v.Len() > *schema.MaxItems {
		return fmt.Errorf("validation error maxItems on %v, expected %v", toString(a), schema)
	}

	if len(schema.Enum) > 0 {
		return checkValueIsInEnum(a, schema.Enum, &Schema{Type: "array", Items: schema.Items})
	}

	if schema.UniqueItems {
		var unique []interface{}
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			for _, u := range unique {
				if compare(item, u) {
					return fmt.Errorf("value %v must contain unique items, expected %v", toString(a), schema)
				}
			}
			unique = append(unique, item)
		}
	}

	return nil
}

func validateObject(i interface{}, schema *Schema) error {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Map {
		if schema.MinProperties != nil && v.Len() < *schema.MinProperties {
			return fmt.Errorf("validation error minProperties on %v, expected %v", toString(i), schema)
		}
		if schema.MaxProperties != nil && v.Len() > *schema.MaxProperties {
			return fmt.Errorf("validation error maxProperties on %v, expected %v", toString(i), schema)
		}
		if !schema.IsFreeForm() && schema.Properties != nil && v.Len() > schema.Properties.Value.Len() {
			return fmt.Errorf("validation error too many fields on %v, expected %v", toString(i), schema)
		}

		for _, p := range schema.Required {
			if e := v.MapIndex(reflect.ValueOf(p)); !e.IsValid() {
				return fmt.Errorf("missing required field %v on %v, expected %v", p, toString(i), schema)
			}
		}
	} else if m, ok := i.(*sortedmap.LinkedHashMap); ok {
		if schema.MinProperties != nil && m.Len() < *schema.MinProperties {
			return fmt.Errorf("validation error minProperties on %v, expected %v", m, schema)
		}
		if schema.MaxProperties != nil && m.Len() > *schema.MaxProperties {
			return fmt.Errorf("validation error maxProperties on %v, expected %v", m, schema)
		}

		if !schema.IsFreeForm() && schema.Properties != nil && m.Len() > schema.Properties.Value.Len() {
			return fmt.Errorf("validation error too many fields on %v, expected %v", toString(i), schema)
		}

		for _, p := range schema.Required {
			if v := m.Get(p); v == nil {
				return fmt.Errorf("missing required field %v on %v, expected %v", p, m, schema)
			}
		}

		if len(schema.Enum) > 0 {
			found := false
		LoopEnum:
			for _, e := range schema.Enum {
				v, ok := e.(map[string]interface{})
				if !ok {
					return fmt.Errorf("expected object in enumeration, got %v", schema.Enum)
				}
				for name, propValue := range v {
					if m.Get(name) != propValue {
						continue LoopEnum
					}
				}
				found = true
			}
			if !found {
				return fmt.Errorf("value '%v' does not match one in the enum %v", m.String(), toString(schema.Enum))
			}
		}
	}

	return nil
}

func checkValueIsInEnum(i interface{}, enum []interface{}, entrySchema *Schema) error {
	found := false
	for _, e := range enum {
		v, err := parse(e, &Ref{Value: entrySchema})
		if err != nil {
			log.Errorf("unable to parse enum value %v to integer: %v", toString(e), err)
			continue
		}
		if compare(i, v) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("value %v does not match one in the enum %v", toString(i), toString(enum))
	}

	return nil
}

func compare(a, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	if av.Kind() != bv.Kind() {
		return false
	}

	switch av.Kind() {
	case reflect.Slice:
		return compareSlice(av, bv)
	default:
		return a == b
	}
}

func compareSlice(a, b reflect.Value) bool {
	if a.Len() != b.Len() {
		return false
	}
	for i := 0; i < a.Len(); i++ {
		if !compare(a.Index(i).Interface(), b.Index(i).Interface()) {
			return false
		}
	}
	return true
}
