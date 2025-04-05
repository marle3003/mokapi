package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/schema"
	"strings"
)

const smallestFloat = 1e-15

func numberNodes() []*Node {
	return []*Node{
		{Name: "number", Fake: fakeNumber},
		{Name: "age", Fake: fakeNumber},
	}
}

func fakeNumber(r *Request) (interface{}, error) {
	s := r.Schema
	minLength := 11
	maxLength := 11
	if s != nil && s.MaxLength != nil {
		maxLength = *s.MaxLength
	}
	if s != nil && s.MinLength != nil {
		minLength = *s.MinLength
	} else if s != nil && s.MaxLength != nil {
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

func fakeAge(r *Request) (interface{}, error) {
	min, max := getRangeWithDefault(0, 50, r.Schema)

	if r.Schema.IsNumber() {
		return gofakeit.Float64Range(min, max), nil
	} else {
		return gofakeit.Number(int(min), int(max)), nil
	}
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
