package schema

import (
	"fmt"
	"github.com/google/uuid"
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

func validateFloat64(n float64, schema *Schema) error {
	if schema.Minimum != nil {
		min := *schema.Minimum
		if schema.ExclusiveMinimum != nil && (*schema.ExclusiveMinimum) && n <= min {
			return fmt.Errorf("%v is lower as the expected minimum %v", n, min)
		} else if n < min {
			return fmt.Errorf("%v is lower as the expected minimum %v", n, min)
		}
	}
	if schema.Maximum != nil {
		max := *schema.Maximum
		if schema.ExclusiveMaximum != nil && (*schema.ExclusiveMaximum) && n >= max {
			return fmt.Errorf("%v is greater as the expected maximum %v", n, max)
		} else if n > max {
			return fmt.Errorf("%v is greater as the expected maximum %v", n, max)
		}
	}
	return nil
}

func validateInt64(n int64, schema *Schema) error {
	if schema.Minimum != nil {
		min := int64(*schema.Minimum)
		if schema.ExclusiveMinimum != nil && (*schema.ExclusiveMinimum) && n <= min {
			return fmt.Errorf("%v is lower as the expected minimum %v", n, min)
		} else if n < min {
			return fmt.Errorf("%v is lower as the expected minimum %v", n, min)
		}
	}
	if schema.Maximum != nil {
		max := int64(*schema.Maximum)
		if schema.ExclusiveMaximum != nil && (*schema.ExclusiveMaximum) && n >= max {
			return fmt.Errorf("%v is greater as the expected maximum %v", n, max)
		} else if n > max {
			return fmt.Errorf("%v is greater as the expected maximum %v", n, max)
		}
	}
	return nil
}
