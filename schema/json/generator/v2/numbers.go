package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"math"
	"mokapi/schema/json/schema"
)

func fakeInteger(s *schema.Schema, min, max int) (interface{}, error) {
	if s.IsAny() {
		return gofakeit.Number(min, max), nil
	}

	if s.Format == "int32" && max > math.MaxInt32 {
		max = math.MaxInt32
	}
	if s.Minimum != nil {
		min = int(*s.Minimum)
	}
	if s.ExclusiveMinimum != nil {
		if s.ExclusiveMinimum.IsA() {
			min = int(s.ExclusiveMinimum.A) + 1
		} else {
			min = int(*s.Minimum) + 1
		}
	}
	if s.Maximum != nil {
		max = int(*s.Maximum)
	}
	if s.ExclusiveMaximum != nil {
		if s.ExclusiveMaximum.IsA() {
			max = int(s.ExclusiveMaximum.A) - 1
		} else {
			max = int(*s.Maximum) - 1
		}

	}

	return gofakeit.Number(min, max), nil
}
