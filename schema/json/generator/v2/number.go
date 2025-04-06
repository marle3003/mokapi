package v2

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"math"
	"mokapi/schema/json/schema"
	"strings"
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
			Fake: fakeNumber,
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
		if s.Format == "int32" {
			return gofakeit.Int32(), nil
		}
		return gofakeit.Int64(), nil
	}
	if hasRange {
		min, max := getRange(s)
		if s.MultipleOf != nil {
			return randomMultiple(int(min), int(max), int(*s.MultipleOf))
		}
		return gofakeit.Number(int(min), int(max)), nil
	}
	min := 0
	max := 10000
	n := gofakeit.Number(min, max)
	return n * int(*s.MultipleOf), nil
}

func fakeIntegerWithRange(s *schema.Schema, min, max int) (any, error) {
	if s.IsAny() {
		return gofakeit.Number(min, max), nil
	}

	minValue, maxValue := getRangeWithDefault(float64(min), float64(max), s)
	min = int(minValue)
	max = int(maxValue)

	if s.MultipleOf != nil {
		return randomMultiple(min, max, int(*s.MultipleOf))

	}
	return gofakeit.Number(min, max), nil
}

func fakeNumber(r *Request) (interface{}, error) {
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
			return gofakeit.Float32(), nil
		}
		return gofakeit.Float64(), nil
	}
	if hasRange {
		min, max := getRange(s)
		if s.MultipleOf != nil {
			return randomFloatMultiple(min, max, *s.MultipleOf)
		}
		return gofakeit.Float64Range(min, max), nil
	}
	min := 0
	max := 10000
	n := gofakeit.Number(min, max)
	return float64(n) * *s.MultipleOf, nil
}

func fakeAge(r *Request) (interface{}, error) {
	min, max := getRangeWithDefault(0, 50, r.Schema)

	if r.Schema.IsNumber() {
		return gofakeit.Float64Range(min, max), nil
	} else {
		return gofakeit.Number(int(min), int(max)), nil
	}
}

func getRange(s *schema.Schema) (float64, float64) {
	var min float64
	var max float64
	if len(s.Type) == 1 && s.Type[0] == "integer" {
		if s.Format == "int32" {
			min = math.MinInt32
			max = math.MaxInt32
		} else {
			min = math.MinInt64
			max = math.MaxInt64
		}
	} else if len(s.Type) == 1 && s.Type[0] == "number" {
		if s.Format == "float" {
			max = math.MaxFloat32
			min = -math.MaxFloat32
		} else {
			max = math.MaxFloat64
			min = -math.MaxFloat64
		}
	}

	return getRangeWithDefault(min, max, s)
}

func getRangeWithDefault(min, max float64, s *schema.Schema) (float64, float64) {
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
