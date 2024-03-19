package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"math"
	"mokapi/json/schema"
	"strings"
)

const smallestFloat = 1e-20

func Numbers() *Tree {
	return &Tree{
		Name: "Numbers",
		Nodes: []*Tree{
			MultipleOf(),
			Id(),
			IdList(),
			Year(),
			Integer32(),
			Integer64(),
			Float32(),
			Number(),
		},
	}
}

func Id() *Tree {
	return &Tree{
		Name: "Id",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(ComparerFunc(func(p *PathElement) bool {
				return (strings.ToLower(p.Name) == "id" || strings.HasSuffix(p.Name, "Id")) &&
					(p.Schema.IsInteger() || p.Schema.IsAny())
			}))
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			return newPositiveNumber(s)
		},
	}
}

func IdList() *Tree {
	return &Tree{
		Name: "ID-List",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(ComparerFunc(func(p *PathElement) bool {
				name := strings.ToLower(p.Name)
				if name == "ids" {
					return p.Schema.IsArray() || p.Schema.IsAny()
				}
				if strings.HasSuffix(name, "ids") {
					return p.Schema.IsArray()
				}

				return false
			}))
		},
		Fake: func(r *Request) (interface{}, error) {
			last := r.Last()
			s := last.Schema
			if s.IsAny() {
				s = &schema.Ref{Value: &schema.Schema{Type: []string{"array"}}}
			}

			next := r.With()
			next.Path = Path{&PathElement{
				Name:   strings.TrimSuffix(last.Name, "s"),
				Schema: s,
			}}
			return r.g.tree.Resolve(next)
		},
	}
}

func Year() *Tree {
	return &Tree{
		Name: "Year",
		Test: func(r *Request) bool {
			last := r.Last()
			return (strings.ToLower(last.Name) == "year" || strings.HasSuffix(last.Name, "Year") ||
				strings.HasPrefix(last.Name, "year")) && (last.Schema.IsInteger() || last.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			if s.IsAny() {
				s = &schema.Schema{Type: []string{"integer"}}
			}
			min, max := getRangeWithDefault(s, 1900, 2199)
			return gofakeit.IntRange(int(min), int(max)), nil
		},
	}
}

func MultipleOf() *Tree {
	return &Tree{
		Name: "MultipleOf",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return (s.IsNumber() || s.IsInteger()) && s.MultipleOf != nil
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			min := 0
			max := 100
			n := gofakeit.Number(min, max)
			v := *s.MultipleOf * float64(n)
			if s.Maximum != nil && v > *s.Maximum {
				for v > *s.Maximum {
					n--
					v = *s.MultipleOf * float64(n)
				}
			}
			return v, nil
		},
	}
}

func Integer32() *Tree {
	return &Tree{
		Name: "Integer32",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s.IsInteger() && s.Format == "int32"
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			if !hasNumberRange(s) {
				return gofakeit.Int32(), nil
			}

			min, max := getRange(s)
			min = math.Ceil(min)
			max = math.Floor(max)

			if err := validateRange(min, max); err != nil {
				return 0, fmt.Errorf("%w in %s", err, s)
			}

			// gofakeit uses Intn function which panics if number is <= 0
			return int32(math.Round(float64(gofakeit.Float32Range(float32(min), float32(max))))), nil
		},
	}
}

func Integer64() *Tree {
	return &Tree{
		Name: "Integer",
		Test: func(r *Request) bool {
			return r.LastSchema().IsInteger()
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			if !hasNumberRange(s) {
				return gofakeit.Int64(), nil
			}

			min, max := getRange(s)
			min = math.Ceil(min)
			max = math.Floor(max)

			if err := validateRange(min, max); err != nil {
				return 0, fmt.Errorf("%w in %s", err, s)
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
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s.IsNumber() && s.Format == "float"
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			if !hasNumberRange(s) {
				return gofakeit.Float32(), nil
			}

			min, max := getRange(s)

			if err := validateRange(min, max); err != nil {
				return 0, fmt.Errorf("%w in %s", err, s)
			}

			return gofakeit.Float32Range(float32(min), float32(max)), nil
		},
	}
}

func Number() *Tree {
	return &Tree{
		Name: "Number",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s.IsNumber()
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			if !hasNumberRange(s) {
				return gofakeit.Float64(), nil
			}

			min, max := getRange(s)

			if err := validateRange(min, max); err != nil {
				return 0, fmt.Errorf("%w in %s", err, s)
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
