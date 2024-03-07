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
			if r.Schema.IsAny() {
				return gofakeit.Number(1, math.MaxInt64), nil
			}
			if r.Schema.IsString() {
				return gofakeit.UUID(), nil
			} else if r.Schema.IsInteger() {
				min := 1
				max := math.MaxInt64

				if r.Schema.Format == "int32" {
					max = math.MaxInt32
				}
				if r.Schema.Minimum != nil {
					min = int(*r.Schema.Minimum)
				}
				if r.Schema.ExclusiveMinimum != nil {
					min = int(*r.Schema.ExclusiveMinimum) + 1
				}
				if r.Schema.Maximum != nil {
					max = int(*r.Schema.Maximum)
				}
				if r.Schema.ExclusiveMaximum != nil {
					max = int(*r.Schema.ExclusiveMaximum) - 1
				}

				return gofakeit.Number(min, max), nil
			}
			return nil, ErrUnsupported
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
