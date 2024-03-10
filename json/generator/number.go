package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"math"
	"mokapi/json/schema"
	"strings"
)

const smallestFloat = 1e-20

func Number() *Tree {
	return &Tree{
		Name: "Number",
		nodes: []*Tree{
			Id(),
			Budget(),
			Price(),
			Integer32(),
			Integer64(),
			Float32(),
			Float64(),
		},
	}
}

func Id() *Tree {
	return &Tree{
		Name: "Id",
		compare: func(r *Request) bool {
			return (r.LastName() == "id" || strings.HasSuffix(r.LastName(), "Id")) &&
				(r.Schema.IsAny() || r.Schema.IsString() || r.Schema.IsInteger())
		},
		resolve: func(r *Request) (interface{}, error) {
			return newPositiveNumber(r.Schema)
		},
	}
}

func Budget() *Tree {
	return &Tree{
		Name: "Budget",
		compare: func(r *Request) bool {
			return (r.LastName() == "budget" || strings.HasSuffix(r.LastName(), "Budget")) &&
				(r.Schema.IsAny() || r.Schema.IsInteger())
		},
		resolve: func(r *Request) (interface{}, error) {
			return newPositiveNumber(r.Schema)
		},
	}
}

func Price() *Tree {
	return &Tree{
		Name: "Price",
		compare: func(r *Request) bool {
			return (r.LastName() == "price" || strings.HasSuffix(r.LastName(), "Price")) &&
				(r.Schema.IsAny() || r.Schema.IsInteger() || r.Schema.IsNumber())
		},
		resolve: func(r *Request) (interface{}, error) {
			s := r.Schema
			if r.Schema.IsAny() {
				s = &schema.Schema{Type: []string{"number"}}
			}
			min, max := getRangeWithDefault(s, 0, 100000)
			return gofakeit.Price(min, max), nil
		},
	}
}

func Integer32() *Tree {
	return &Tree{
		Name: "Integer32",
		compare: func(r *Request) bool {
			return r.Schema.IsInteger() && r.Schema.Format == "int32"
		},
		resolve: func(r *Request) (interface{}, error) {
			if !hasNumberRange(r.Schema) {
				return gofakeit.Int32(), nil
			}

			min, max := getRange(r.Schema)
			min = math.Ceil(min)
			max = math.Floor(max)

			if err := validateRange(min, max); err != nil {
				return 0, fmt.Errorf("%w in %s", err, r.Schema)
			}
			if max-min > math.MaxInt32 {
				return gofakeit.Int32(), nil
			}

			// gofakeit uses Intn function which panics if number is <= 0
			return int32(math.Round(float64(gofakeit.Float32Range(float32(min), float32(max))))), nil
		},
	}
}

func Integer64() *Tree {
	return &Tree{
		Name: "Integer64",
		compare: func(r *Request) bool {
			return r.Schema.IsInteger()
		},
		resolve: func(r *Request) (interface{}, error) {
			if !hasNumberRange(r.Schema) {
				return gofakeit.Int64(), nil
			}

			min, max := getRange(r.Schema)
			min = math.Ceil(min)
			max = math.Floor(max)

			if err := validateRange(min, max); err != nil {
				return 0, fmt.Errorf("%w in %s", err, r.Schema)
			}
			if math.IsInf(max-min, 0) {
				return gofakeit.Int64(), nil
			}

			// gofakeit uses Intn function which panics if number is <= 0
			return int64(math.Round(gofakeit.Float64Range(min, max))), nil
		},
	}
}

func Float32() *Tree {
	return &Tree{
		Name: "Float32",
		compare: func(r *Request) bool {
			return r.Schema.IsNumber() && r.Schema.Format == "float"
		},
		resolve: func(r *Request) (interface{}, error) {
			if !hasNumberRange(r.Schema) {
				return gofakeit.Float32(), nil
			}

			min, max := getRange(r.Schema)

			if err := validateRange(min, max); err != nil {
				return 0, fmt.Errorf("%w in %s", err, r.Schema)
			}
			if max-min > math.MaxFloat32 {
				return gofakeit.Float32(), nil
			}

			return gofakeit.Float32Range(float32(min), float32(max)), nil
		},
	}
}

func Float64() *Tree {
	return &Tree{
		Name: "Float64",
		compare: func(r *Request) bool {
			return r.Schema.IsNumber()
		},
		resolve: func(r *Request) (interface{}, error) {
			if !hasNumberRange(r.Schema) {
				return gofakeit.Float64(), nil
			}

			min, max := getRange(r.Schema)

			if err := validateRange(min, max); err != nil {
				return 0, fmt.Errorf("%w in %s", err, r.Schema)
			}
			if math.IsInf(max-min, 0) {
				return gofakeit.Float64(), nil
			}

			return gofakeit.Float64Range(min, max), nil
		},
	}
}

func validateRange(min, max float64) error {
	if min >= max {
		return fmt.Errorf("invalid minimum '%v' and maximum '%v'", min, max)
	}
	return nil
}

func hasNumberRange(s *schema.Schema) bool {
	return s.Minimum != nil || s.Maximum != nil || s.ExclusiveMinimum != nil || s.ExclusiveMaximum != nil
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
			min = max * -1
		} else {
			max = math.MaxFloat64
			min = max * -1
		}
	}

	return getRangeWithDefault(s, min, max)
}

func getRangeWithDefault(s *schema.Schema, min, max float64) (float64, float64) {
	if s.Minimum != nil {
		min = *s.Minimum
	}
	if s.Maximum != nil {
		max = *s.Maximum
	}
	if s.ExclusiveMinimum != nil {
		min = *s.ExclusiveMinimum + smallestFloat
	}
	if s.ExclusiveMaximum != nil {
		max = *s.ExclusiveMaximum - smallestFloat
	}

	return min, max
}

func newPositiveNumber(s *schema.Schema) (interface{}, error) {
	if s.IsAny() {
		return gofakeit.Number(1, math.MaxInt64), nil
	}
	if s.IsString() {
		return gofakeit.UUID(), nil
	} else if s.IsInteger() {
		min := 1
		max := math.MaxInt64

		if s.Format == "int32" {
			max = math.MaxInt32
		}
		if s.Minimum != nil {
			min = int(*s.Minimum)
		}
		if s.ExclusiveMinimum != nil {
			min = int(*s.ExclusiveMinimum) + 1
		}
		if s.Maximum != nil {
			max = int(*s.Maximum)
		}
		if s.ExclusiveMaximum != nil {
			max = int(*s.ExclusiveMaximum) - 1
		}

		return gofakeit.Number(min, max), nil
	}
	return nil, ErrUnsupported
}
