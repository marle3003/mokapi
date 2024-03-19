package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
	"strconv"
	"strings"
)

func Addresses() *Tree {
	root := &Tree{
		Name: "Address",
		Nodes: []*Tree{
			CoAddress(),
			City(),
			Country(),
			//Postcodes(),
			Postcode(),
			Longitude(),
			Latitude(),
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

func Country() *Tree {
	return &Tree{
		Name: "Country",
		Test: func(r *Request) bool {
			last := r.Last()
			return last.Name == "country" && (last.Schema.IsAny() || last.Schema.IsString())
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			max := 2
			if s != nil && s.MaxLength != nil {
				max = *s.MaxLength
			}

			if max == 2 {
				return gofakeit.CountryAbr(), nil
			}

			return gofakeit.Country(), nil
		},
	}
}

//
//func Postcodes() *Tree {
//	return &Tree{
//		Name: "Postcodes",
//		Test: func(r *Request) bool {
//			last := strings.ToLower(r.LastName())
//			return (last == "postcodes" || last == "zips") &&
//				(r.Schema.IsArray() || r.Schema.IsAny())
//		},
//		Fake: func(r *Request) (interface{}, error) {
//			next := r.With(Name("postcode"))
//			if r.Schema.IsAny() {
//				next = next.With(Schema(&schema.Schema{Type: []string{"array"}}))
//			}
//			return r.g.tree.Resolve(next)
//		},
//	}
//}

func Postcode() *Tree {
	return &Tree{
		Name: "Postcode",
		Test: func(r *Request) bool {
			last := r.Last()
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

func Longitude() *Tree {
	return &Tree{
		Name: "Longitude",
		Test: func(r *Request) bool {
			last := r.Last()
			name := strings.ToLower(last.Name)
			return name == "longitude" && (last.Schema.IsNumber() || last.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Longitude(), nil
		},
	}
}

func Latitude() *Tree {
	return &Tree{
		Name: "Latitude",
		Test: func(r *Request) bool {
			last := r.Last()
			name := strings.ToLower(last.Name)
			return name == "latitude" && (last.Schema.IsNumber() || last.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Latitude(), nil
		},
	}
}

func Street() *Tree {
	return &Tree{
		Name: "Street",
		Test: func(r *Request) bool {
			last := r.Last()
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
	min := 4
	max := 6
	if s.IsInteger() {
		if s.Minimum != nil {
			min = len(fmt.Sprintf("%v", *s.Minimum))
		}
		if s.Maximum != nil {
			max = len(fmt.Sprintf("%v", *s.Maximum))
		}
	} else if s.IsString() {
		if s.MinLength != nil {
			min = *s.MinLength
		}
		if s.MaxLength != nil {
			max = *s.MaxLength
		}
	}

	var n int
	if min == max {
		n = min
	} else {
		n = gofakeit.Number(min, max)
	}

	code := gofakeit.Numerify(strings.Repeat("#", n))
	if s.IsInteger() {
		codeN, _ := strconv.ParseInt(code, 10, 32)
		return int(codeN)
	}
	return code
}
