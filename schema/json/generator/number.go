package generator

import (
	"fmt"
	"math"
	"mokapi/schema/json/schema"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
)

const (
	defaultMin = -1e6
	defaultMax = 1e6
)

const smallestFloat = 1e-15

func numbers() []*Node {
	return []*Node{
		{
			Name: "number",
			Fake: func(r *Request) (any, error) {
				if r.Schema == nil {
					r = r.WithSchema(&schema.Schema{Type: schema.Types{"string"}})
				}
				return fakeNumber(r)
			},
		},
		{
			Name: "age",
			Fake: fakeAge,
		},
		{
			Name: "year",
			Fake: fakeYear,
		},
		{
			Name: "quantity",
			Fake: func(r *Request) (interface{}, error) {
				return fakeIntegerWithRange(r.Schema, 0, 100)
			},
		},
	}
}

func fakeYear(r *Request) (interface{}, error) {
	s := r.Schema
	if s.IsAny() {
		s = &schema.Schema{Type: []string{"integer"}}
	}
	return fakeIntegerWithRange(s, 1900, 2199)
}

func fakeInteger(s *schema.Schema) (any, error) {
	hasRange := hasNumberRange(s)
	if !hasRange && s.MultipleOf == nil {
		return int64(gofakeit.Number(defaultMin, defaultMax)), nil
	}
	if hasRange {
		minValue, maxValue := getRangeWithDefault(s, defaultMin, defaultMax)
		if err := validateRange(minValue, maxValue); err != nil {
			return nil, fmt.Errorf("%w in %s", err, s)
		}
		if s.MultipleOf != nil {
			v, err := randomMultiple(int(minValue), int(maxValue), int(*s.MultipleOf))
			if err != nil {
				return nil, err
			}
			return int64(v), nil
		}
		v := int64(math.Round(gofakeit.Float64Range(minValue, maxValue)))
		return v, nil
	}
	minValue := 0
	maxValue := 100
	n := gofakeit.Number(minValue, maxValue)
	v := n * int(*s.MultipleOf)
	return int64(v), nil
}

func fakeIntegerWithRange(s *schema.Schema, min, max int) (any, error) {
	if s.IsAny() {
		return gofakeit.Number(min, max), nil
	}

	minValue, maxValue := getRangeWithDefault(s, float64(min), float64(max))
	if err := validateRange(minValue, maxValue); err != nil {
		return nil, fmt.Errorf("%w in %s", err, s)
	}

	min = int(minValue)
	max = int(maxValue)

	if s.MultipleOf != nil {
		v, err := randomMultiple(min, max, int(*s.MultipleOf))
		if err != nil {
			return nil, err
		}
		if s.Format == "int32" {
			return int32(v), nil
		}
		return int64(v), nil
	}
	v := gofakeit.Number(min, max)
	if s.Format == "int32" {
		return int32(v), nil
	}
	return int64(v), nil
}

func fakeNumber(r *Request) (interface{}, error) {
	n, err := newNumber(r)
	if err != nil {
		return nil, err
	}
	if shouldEnsureDecimalPart(r.Schema) && isInteger(n) {
		switch n.(type) {
		case float64:
			frac := r.g.rand.Float64()
			return n.(float64) + frac, nil
		case float32:
			frac := r.g.rand.Float32()
			return n.(float32) + frac, nil
		}
	}
	return n, nil
}

func newNumber(r *Request) (interface{}, error) {
	s := r.Schema
	if s == nil {
		return gofakeit.Float64(), nil
	}

	if s.IsString() {
		minLength := 11
		maxLength := 11
		if s.MaxLength != nil {
			maxLength = *s.MaxLength
		}
		if s.MinLength != nil {
			minLength = *s.MinLength
		} else if s.MaxLength != nil {
			minLength = 0
		}
		var n int
		if minLength == maxLength {
			n = minLength
		} else {
			n = gofakeit.Number(minLength, maxLength)
		}
		return gofakeit.Numerify(strings.Repeat("#", n)), nil
	}
	if s.IsInteger() {
		return fakeInteger(s)
	}

	hasRange := hasNumberRange(s)
	if !hasRange && s.MultipleOf == nil {
		if s.Format == "float" {
			return float32(gofakeit.Float64Range(defaultMin, defaultMax)), nil
		}
		return gofakeit.Float64Range(defaultMin, defaultMax), nil
	}
	if hasRange {
		minValue, maxValue := getRangeWithDefault(s, defaultMin, defaultMax)
		if err := validateRange(minValue, maxValue); err != nil {
			return nil, fmt.Errorf("%w in %s", err, s)
		}
		if s.MultipleOf != nil {
			v, err := randomFloatMultiple(minValue, maxValue, *s.MultipleOf)
			if err != nil {
				return nil, err
			}
			if s.Format == "float" {
				return float32(v), nil
			}
			return v, nil
		}
		v := gofakeit.Float64Range(minValue, maxValue)
		if s.Format == "float" {
			return float32(v), nil
		}
		return v, nil
	}
	minValue := 0
	maxValue := 100
	n := gofakeit.Number(minValue, maxValue)
	v := float64(n) * *s.MultipleOf
	if s.Format == "float" {
		return float32(v), nil
	}
	return v, nil
}

func fakeAge(r *Request) (interface{}, error) {
	minValue, maxValue := getRangeWithDefault(r.Schema, 0, 100)
	return int64(gofakeit.Number(int(minValue), int(maxValue))), nil
}

func getRangeWithDefault(s *schema.Schema, min, max float64) (float64, float64) {
	if s == nil {
		return min, max
	}

	if s.Minimum != nil {
		min = *s.Minimum
	}
	if s.Maximum != nil {
		max = *s.Maximum
	}

	modifier := smallestFloat
	if s.IsInteger() {
		modifier = 1
	}

	if s.ExclusiveMinimum != nil {
		if s.ExclusiveMinimum.IsA() {
			min = s.ExclusiveMinimum.A + modifier
		} else {
			min = *s.Minimum + modifier
		}
	}
	if s.ExclusiveMaximum != nil {
		if s.ExclusiveMaximum.IsA() {
			max = s.ExclusiveMaximum.A - modifier
		} else {
			max = *s.Maximum - modifier
		}
	}

	return min, max
}

func hasNumberRange(s *schema.Schema) bool {
	return s.Minimum != nil || s.Maximum != nil || s.ExclusiveMinimum != nil || s.ExclusiveMaximum != nil
}

func randomMultiple(min, max, multipleOf int) (int, error) {
	// Adjust min and max to be aligned with multipleOf
	adjustedMin := ((min + multipleOf - 1) / multipleOf) * multipleOf
	adjustedMax := (max / multipleOf) * multipleOf

	if adjustedMin > adjustedMax {
		return 0, fmt.Errorf("no valid multiple in range")
	}

	count := ((adjustedMax - adjustedMin) / multipleOf) + 1
	n := gofakeit.Number(0, count)

	return adjustedMin + n*multipleOf, nil
}

func randomFloatMultiple(min, max, multipleOf float64) (float64, error) {
	start := math.Ceil(min / multipleOf)
	end := math.Floor(max / multipleOf)

	if start > end {
		return 0, fmt.Errorf("no valid multiple in range")
	}

	n := gofakeit.Number(0, int(end-start)+1) + int(start)
	return float64(n) * multipleOf, nil
}

func validateRange(min, max float64) error {
	if min > max {
		return fmt.Errorf("invalid minimum '%v' and maximum '%v'", min, max)
	}
	return nil
}

func shouldEnsureDecimalPart(s *schema.Schema) bool {
	if s == nil || s.Not == nil {
		return false
	}
	if s.Not.IsInteger() && schemaNoConstraintsForType(s.Not, "integer") {
		return true
	}
	if s.Not.AnyOf == nil {
		return false
	}
	for _, as := range s.Not.AnyOf {
		if shouldEnsureDecimalPart(as) {
			return true
		}
	}
	return false
}

func isInteger(n any) bool {
	switch t := n.(type) {
	case int, int64:
		return true
	case float64:
		_, frac := math.Modf(t)
		// Some values that look like integers may not be stored exactly as such (e.g. 2.0000000000000004).
		return math.Abs(frac) < 1e-9
	case float32:
		_, frac := math.Modf(float64(t))
		return math.Abs(frac) > 1e-6 // slightly looser than float64
	default:
		return false
	}
}
