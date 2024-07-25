package parser

import (
	"fmt"
	"math"
	"mokapi/schema/json/schema"
	"strconv"
)

func (p *Parser) ParseInteger(i interface{}, s *schema.Schema) (n int64, err error) {
	switch v := i.(type) {
	case int:
		n = int64(v)
	case int64:
		n = v
	case float64:
		if math.Trunc(v) != v {
			return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
		}
		n = int64(v)
	case int32:
		n = int64(v)
	case string:
		if !p.ConvertStringToNumber {
			return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
		}
		switch s.Format {
		case "int64":
			n, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
			}
			return n, nil
		default:
			temp, err := strconv.Atoi(v)
			if err != nil {
				return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
			}
			n = int64(temp)
		}
	default:
		return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
	}

	switch s.Format {
	case "int32":
		if n > math.MaxInt32 || n < math.MinInt32 {
			return 0, fmt.Errorf("parse '%v' failed: represents a number either less than int32 min value or greater max value, expected %v", i, s)
		}
	}

	if s.MultipleOf != nil {
		m := float64(n) / *s.MultipleOf
		if m != float64(int(m)) {
			return 0, fmt.Errorf("%v is not a multiple of %v: %v", n, *s.MultipleOf, s)
		}
	}

	return n, validateInt64(n, s)
}

func validateInt64(n int64, s *schema.Schema) error {
	if s.ExclusiveMinimum != nil && (s.ExclusiveMinimum.IsA() || s.ExclusiveMinimum.B) {
		var min int64
		if s.ExclusiveMinimum.IsA() {
			min = int64(s.ExclusiveMinimum.A)
		} else {
			if s.Minimum == nil {
				return fmt.Errorf("exclusiveMinimum is set to true but no minimum value is specified")
			}
			min = int64(*s.Minimum)
		}
		if n <= min {
			return fmt.Errorf("%v is lower or equal as the required minimum %v, expected %v", n, min, s)
		}
	} else if s.Minimum != nil && n < int64(*s.Minimum) {
		return fmt.Errorf("%v is lower as the required minimum %v, expected %v", n, *s.Minimum, s)
	}

	if s.ExclusiveMaximum != nil && (s.ExclusiveMaximum.IsA() || s.ExclusiveMaximum.B) {
		var max int64
		if s.ExclusiveMaximum.IsA() {
			max = int64(s.ExclusiveMaximum.A)
		} else {
			if s.Maximum == nil {
				return fmt.Errorf("exclusiveMaximum is set to true but no maximum value is specified")
			}
			max = int64(*s.Maximum)
		}
		if n >= max {
			return fmt.Errorf("%v is greater or equal as the required maximum %v, expected %v", n, max, s)
		}
	} else if s.Maximum != nil && n > int64(*s.Maximum) {
		return fmt.Errorf("%v is greater as the required maximum %v, expected %v", n, *s.Maximum, s)
	}

	if len(s.Enum) > 0 {
		return checkValueIsInEnum(n, s.Enum, &schema.Schema{Type: schema.Types{"integer"}, Format: "int64"})
	}

	return nil
}
