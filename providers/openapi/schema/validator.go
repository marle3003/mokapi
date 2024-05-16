package schema

import (
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	jsonSchema "mokapi/json/schema"
	"mokapi/sortedmap"
	"net"
	"net/mail"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
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
		return checkValueIsInEnum(str, s.Enum, &Schema{Type: jsonSchema.Types{"string"}})
	}

	return nil
}

func validateFloat64(n float64, schema *Schema) error {
	if schema.ExclusiveMinimum != nil && (schema.ExclusiveMinimum.IsA() || schema.ExclusiveMinimum.B) {
		min := 0.0
		if schema.ExclusiveMinimum.IsA() {
			min = schema.ExclusiveMinimum.A
		} else if schema.Minimum != nil {
			min = *schema.Minimum
		} else {
			return fmt.Errorf("exclusiveMinimum is set to true but no minimum is set")
		}
		if n <= min {
			return fmt.Errorf("%v is lower or equal as the required minimum %v, expected %v", n, min, schema)
		}
	} else if schema.Minimum != nil && n < *schema.Minimum {
		return fmt.Errorf("%v is lower as the required minimum %v, expected %v", n, *schema.Minimum, schema)
	}

	if schema.ExclusiveMaximum != nil && (schema.ExclusiveMaximum.IsA() || schema.ExclusiveMaximum.B) {
		max := 0.0
		if schema.ExclusiveMaximum.IsA() {
			max = schema.ExclusiveMaximum.A
		} else if schema.Maximum != nil {
			max = *schema.Maximum
		} else {
			return fmt.Errorf("exclusiveMaximum is set to true but no maximum is set")
		}
		if n >= max {
			return fmt.Errorf("%v is greater or equal as the required maximum %v, expected %v", n, max, schema)
		}
	} else if schema.Maximum != nil && n > *schema.Maximum {
		return fmt.Errorf("%v is greater as the required maximum %v, expected %v", n, *schema.Maximum, schema)
	}

	if len(schema.Enum) > 0 {
		return checkValueIsInEnum(n, schema.Enum, &Schema{Type: jsonSchema.Types{"number"}, Format: "double"})
	}

	return nil
}

func validateInt64(n int64, schema *Schema) error {
	if schema.ExclusiveMinimum != nil && (schema.ExclusiveMinimum.IsA() || schema.ExclusiveMinimum.B) {
		min := int64(0)
		if schema.ExclusiveMinimum.IsA() {
			min = int64(schema.ExclusiveMinimum.A)
		} else if schema.Minimum != nil {
			min = int64(*schema.Minimum)
		} else {
			return fmt.Errorf("exclusiveMinimum is set to true but no minimum is set")
		}
		if n <= min {
			return fmt.Errorf("%v is lower or equal as the required minimum %v, expected %v", n, min, schema)
		}
	} else if schema.Minimum != nil && n < int64(*schema.Minimum) {
		return fmt.Errorf("%v is lower as the required minimum %v, expected %v", n, *schema.Minimum, schema)
	}

	if schema.ExclusiveMaximum != nil && (schema.ExclusiveMaximum.IsA() || schema.ExclusiveMaximum.B) {
		max := int64(0)
		if schema.ExclusiveMaximum.IsA() {
			max = int64(schema.ExclusiveMaximum.A)
		} else if schema.Maximum != nil {
			max = int64(*schema.Maximum)
		} else {
			return fmt.Errorf("exclusiveMaximum is set to true but no maximum is set")
		}
		if n >= max {
			return fmt.Errorf("%v is greater or equal as the required maximum %v, expected %v", n, max, schema)
		}
	} else if schema.Maximum != nil && n > int64(*schema.Maximum) {
		return fmt.Errorf("%v is greater as the required maximum %v, expected %v", n, *schema.Maximum, schema)
	}

	if len(schema.Enum) > 0 {
		return checkValueIsInEnum(n, schema.Enum, &Schema{Type: jsonSchema.Types{"integer"}, Format: "int64"})
	}

	return nil
}

func validateArray(a interface{}, schema *Schema) error {
	v := reflect.ValueOf(a)
	if schema.MinItems != nil && v.Len() < *schema.MinItems {
		return fmt.Errorf("should NOT have less than %v items", *schema.MinItems)
	}
	if schema.MaxItems != nil && v.Len() > *schema.MaxItems {
		return fmt.Errorf("should NOT have more than %v items", *schema.MaxItems)
	}

	if len(schema.Enum) > 0 {
		return checkValueIsInEnum(a, schema.Enum, &Schema{Type: jsonSchema.Types{"array"}, Items: schema.Items})
	}

	if schema.UniqueItems {
		var unique []interface{}
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			for _, u := range unique {
				if compare(item, u) {
					return fmt.Errorf("should NOT have duplicate items (%v)", toString(item))
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
		if !schema.IsFreeForm() && schema.Properties != nil {
			var add []string
			for _, k := range v.MapKeys() {
				name := k.Interface().(string)
				if r := schema.Properties.Get(name); r == nil {
					add = append(add, name)
				}
			}
			if len(add) > 0 {
				sort.Strings(add)
				return fmt.Errorf("additional properties not allowed: %v, expected %v", strings.Join(add, ", "), schema)
			}
		}

		for _, p := range schema.Required {
			if e := v.MapIndex(reflect.ValueOf(p)); !e.IsValid() {
				return fmt.Errorf("missing required field '%v'", p)
			}
		}
	} else if m, ok := i.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
		if schema.MinProperties != nil && m.Len() < *schema.MinProperties {
			return fmt.Errorf("validation error minProperties on %v, expected %v", m, schema)
		}
		if schema.MaxProperties != nil && m.Len() > *schema.MaxProperties {
			return fmt.Errorf("validation error maxProperties on %v, expected %v", m, schema)
		}

		if !schema.IsFreeForm() && schema.Properties != nil {
			var add []string
			for it := m.Iter(); it.Next(); {
				name := it.Key()
				if r := schema.Properties.Get(name); r == nil {
					add = append(add, name)
				}
			}
			if len(add) > 0 {
				sort.Strings(add)
				return fmt.Errorf("additional properties not allowed: %v, expected %v", strings.Join(add, ", "), schema)
			}
		}

		for _, p := range schema.Required {
			if _, found := m.Get(p); !found {
				return fmt.Errorf("missing required field '%v'", p)
			}
		}
	}

	if len(schema.Enum) > 0 {
		return checkValueIsInEnum(i, schema.Enum, &Schema{Type: jsonSchema.Types{"object"}})
	}

	return nil
}

func checkValueIsInEnum(i interface{}, enum []interface{}, entrySchema *Schema) error {
	found := false
	p := parser{}
	for _, e := range enum {
		v, err := p.parse(e, &Ref{Value: entrySchema})
		if err != nil {
			log.Errorf("unable to parse enum value %v to %v: %v", toString(e), entrySchema, err)
			continue
		}
		if compare(i, v) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("value '%v' does not match one in the enumeration %v", toString(i), toString(enum))
	}

	return nil
}

func compare(a, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	if av.Kind() != bv.Kind() {
		return false
	}

	k := av.Kind()
	_ = k

	switch av.Kind() {
	case reflect.Slice:
		return compareSlice(av, bv)
	case reflect.Map:
		return compareMap(av, bv)
	case reflect.Struct, reflect.Pointer:
		return compareStruct(av, bv)
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

func compareMap(a, b reflect.Value) bool {
	if a.Len() != b.Len() {
		return false
	}
	for _, k := range a.MapKeys() {
		av := a.MapIndex(k)
		bv := b.MapIndex(k)
		if !compare(av.Interface(), bv.Interface()) {
			return false
		}

	}

	return true
}

func compareStruct(a, b reflect.Value) bool {
	if b.Kind() == reflect.Pointer {
		b = b.Elem()
	}

	m := a.Interface().(*sortedmap.LinkedHashMap[string, interface{}])
	for it := m.Iter(); it.Next(); {
		name := toFieldName(it.Key())
		v := b.FieldByName(name)
		if !compare(it.Value(), v.Interface()) {
			return false
		}
	}
	return true
}

func toFieldName(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
