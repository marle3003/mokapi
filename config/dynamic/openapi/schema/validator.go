package schema

import (
	"fmt"
	"github.com/google/uuid"
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

	return nil
}

func validateFloat64(n float64, schema *Schema) error {
	if schema.Minimum != nil {
		min := *schema.Minimum
		if schema.ExclusiveMinimum != nil && (*schema.ExclusiveMinimum) && n <= min {
			return fmt.Errorf("%v is lower as the required minimum %v, expected %v", n, min, schema)
		} else if n < min {
			return fmt.Errorf("%v is lower as the required minimum %v, expected %v", n, min, schema)
		}
	}
	if schema.Maximum != nil {
		max := *schema.Maximum
		if schema.ExclusiveMaximum != nil && (*schema.ExclusiveMaximum) && n >= max {
			return fmt.Errorf("%v is greater as the required maximum %v, expected %v", n, max, schema)
		} else if n > max {
			return fmt.Errorf("%v is greater as the required maximum %v, expected %v", n, max, schema)
		}
	}
	return nil
}

func validateInt64(n int64, schema *Schema) error {
	if schema.Minimum != nil {
		min := int64(*schema.Minimum)
		if schema.ExclusiveMinimum != nil && (*schema.ExclusiveMinimum) && n <= min {
			return fmt.Errorf("%v is lower as the required minimum %v, expected %v", n, min, schema)
		} else if n < min {
			return fmt.Errorf("%v is lower as the required minimum %v, expected %v", n, min, schema)
		}
	}
	if schema.Maximum != nil {
		max := int64(*schema.Maximum)
		if schema.ExclusiveMaximum != nil && (*schema.ExclusiveMaximum) && n >= max {
			return fmt.Errorf("%v is greater as the required maximum %v, expected %v", n, max, schema)
		} else if n > max {
			return fmt.Errorf("%v is greater as the required maximum %v, expected %v", n, max, schema)
		}
	}
	return nil
}

func validateArray(a interface{}, schema *Schema) error {
	v := reflect.ValueOf(a)
	if schema.MinItems != nil && v.Len() < *schema.MinItems {
		return fmt.Errorf("validation error minItems on %v, expected %v", v.Interface(), schema)
	}
	if schema.MaxItems != nil && v.Len() > *schema.MaxItems {
		return fmt.Errorf("validation error maxItems on %v, expected %v", v.Interface(), schema)
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
				return fmt.Errorf("missing required field %v in %v, expected: %v", p, m, schema)
			}
		}
	}

	return nil
}
