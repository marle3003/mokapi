package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/schema"
	"strconv"
	"strings"
)

func Address() *Tree {
	root := &Tree{
		Name: "Address",
		Nodes: []*Tree{
			CoAddress(),
			Postcode(),
			Street(),
			OpenAddress(),
			AnyAddress(),
		},
	}
	return root
}

func CoAddress() *Tree {
	return &Tree{
		Name: "CoAddress",
		Test: func(r *Request) bool {
			last := r.Last()
			if last == nil {
				return false
			}
			return strings.ToLower(last.Name) == "coaddress" && last.Schema.IsString()
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Name(), nil
		},
	}
}

func AnyAddress() *Tree {
	return &Tree{
		Name: "AnyAddress",
		Test: func(r *Request) bool {
			last := r.Last()
			if last == nil {
				return false
			}
			return strings.ToLower(last.Name) == "address" && last.Schema.IsAny()
		},
		Fake: func(r *Request) (interface{}, error) {
			addr := gofakeit.Address()
			return map[string]interface{}{
				"address":   addr.Address,
				"street":    addr.Street,
				"city":      addr.City,
				"state":     addr.State,
				"zip":       addr.Zip,
				"country":   addr.Country,
				"latitude":  addr.Latitude,
				"longitude": addr.Longitude,
			}, nil
		},
	}
}

func City() *Tree {
	return &Tree{
		Name: "City",
		Test: func(r *Request) bool {
			last := r.Last()
			if last == nil {
				return false
			}
			return last.Name == "city" && (last.Schema.IsAny() || last.Schema.IsString() || last.Schema.IsInteger())
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			if s.IsAny() || s.IsString() {
				return gofakeit.City(), nil
			} else if s.IsInteger() {
				return newPostCode(s), nil
			}
			return nil, ErrUnsupported
		},
	}
}

func Postcode() *Tree {
	return &Tree{
		Name: "Postcode",
		Test: func(r *Request) bool {
			last := r.Last()
			if last == nil {
				return false
			}
			name := strings.ToLower(last.Name)
			s := last.Schema
			return (name == "postcode" || name == "zip") &&
				(s.IsNumber() || s.IsInteger() || s.IsString() || s.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			return newPostCode(r.LastSchema()), nil
		},
	}
}

func Street() *Tree {
	return &Tree{
		Name: "Street",
		Test: func(r *Request) bool {
			last := r.Last()
			if last == nil {
				return false
			}
			name := strings.ToLower(last.Name)
			return name == "street" && last.Schema.IsString()
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Street(), nil
		},
	}
}

func OpenAddress() *Tree {
	return &Tree{
		Name: "OpenAddress",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(ComparerFunc(func(p *PathElement) bool {
				return strings.ToLower(p.Name) == "address" || strings.HasSuffix(p.Name, "Address")
			}), Any())
		},
		Nodes: []*Tree{
			{
				Name: "Line1",
				Test: func(r *Request) bool {
					return strings.ToLower(r.LastName()) == "line1"
				},
				Fake: func(r *Request) (interface{}, error) {
					return gofakeit.Name(), nil
				},
			},
			{
				Name: "Line2",
				Test: func(r *Request) bool {
					return strings.ToLower(r.LastName()) == "line2"
				},
				Fake: func(r *Request) (interface{}, error) {
					return gofakeit.Street(), nil
				},
			},
			{
				Name: "Line3",
				Test: func(r *Request) bool {
					return strings.ToLower(r.LastName()) == "line3"
				},
				Fake: func(r *Request) (interface{}, error) {
					return fmt.Sprintf("%v %v %v", gofakeit.City(), gofakeit.StateAbr(), gofakeit.Zip()), nil
				},
			},
		},
	}
}

func newPostCode(s *schema.Schema) interface{} {
	if s == nil || s.IsAny() {
		s = &schema.Schema{Type: []string{"string"}}
	}
	minLength := 4
	maxLength := 6
	if s.IsInteger() {
		if s.Minimum != nil {
			minLength = len(fmt.Sprintf("%v", *s.Minimum))
		}
		if s.Maximum != nil {
			maxLength = len(fmt.Sprintf("%v", *s.Maximum))
		}
	} else if s.IsString() {
		if s.MinLength != nil {
			minLength = *s.MinLength
		}
		if s.MaxLength != nil {
			maxLength = *s.MaxLength
		}
	}

	var n int
	if minLength == maxLength {
		n = minLength
	} else {
		n = gofakeit.Number(minLength, maxLength)
	}

	code := gofakeit.Numerify(strings.Repeat("#", n))
	if s.IsInteger() {
		codeN, _ := strconv.ParseInt(code, 10, 32)
		return int(codeN)
	} else if s.IsNumber() {
		codeN, _ := strconv.ParseInt(code, 10, 32)
		return float64(codeN)
	}
	return code
}
