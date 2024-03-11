package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
	"strconv"
	"strings"
)

func AddressTree() *Tree {
	root := &Tree{
		Name: "Address",
		nodes: []*Tree{
			City(),
			Cities(),
			Country(),
			Postcodes(),
			Postcode(),
			Longitude(),
			Latitude(),
			CoAddress(),
			Street(),
			OpenAddress(),
			AnyAddress(),
		},
	}
	return root
}

func CoAddress() *Tree {
	return &Tree{
		Name: "City",
		compare: func(r *Request) bool {
			last := strings.ToLower(r.LastName())
			return last == "coaddress" && r.Schema.IsString()
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.Name(), nil
		},
	}
}

func AnyAddress() *Tree {
	return &Tree{
		Name: "AnyAddress",
		compare: func(r *Request) bool {
			last := strings.ToLower(r.LastName())
			return last == "address" && r.Schema.IsAny()
		},
		resolve: func(r *Request) (interface{}, error) {
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
		compare: func(r *Request) bool {
			return r.LastName() == "city" && (r.Schema.IsAny() || r.Schema.IsString() || r.Schema.IsInteger())
		},
		resolve: func(r *Request) (interface{}, error) {
			if r.Schema.IsAny() || r.Schema.IsString() {
				return gofakeit.City(), nil
			} else if r.Schema.IsInteger() {
				return gofakeit.Zip(), nil
			}
			return nil, ErrUnsupported
		},
	}
}

func Cities() *Tree {
	return &Tree{
		Name: "Cities",
		compare: func(r *Request) bool {
			return r.LastName() == "cities" && (r.Schema.IsAny() || r.Schema.IsArray())
		},
		resolve: func(r *Request) (interface{}, error) {
			return r.g.tree.Resolve(r.With(Name("city")))
		},
	}
}

func Country() *Tree {
	return &Tree{
		Name: "Country",
		compare: func(r *Request) bool {
			return r.LastName() == "country" && (r.Schema.IsAny() || r.Schema.IsString())
		},
		resolve: func(r *Request) (interface{}, error) {
			address := r.GetName(-2)
			if strings.HasSuffix(address, "Address") || strings.ToLower(address) == "address" {
				return gofakeit.CountryAbr(), nil
			}
			return gofakeit.Country(), nil
		},
	}
}

func Postcodes() *Tree {
	return &Tree{
		Name: "Postcodes",
		compare: func(r *Request) bool {
			last := strings.ToLower(r.LastName())
			return (last == "postcodes" || last == "zips") &&
				(r.Schema.IsArray() || r.Schema.IsAny())
		},
		resolve: func(r *Request) (interface{}, error) {
			next := r.With(Name("postcode"))
			if r.Schema.IsAny() {
				next = next.With(Schema(&schema.Schema{Type: []string{"array"}}))
			}
			return r.g.tree.Resolve(next)
		},
	}
}

func Postcode() *Tree {
	return &Tree{
		Name: "Postcode",
		compare: func(r *Request) bool {
			last := strings.ToLower(r.LastName())
			return (last == "postcode" || last == "zip") &&
				(r.Schema.IsNumber() || r.Schema.IsInteger() || r.Schema.IsString() || r.Schema.IsAny())
		},
		resolve: func(r *Request) (interface{}, error) {
			return newPostCode(r.Schema), nil
		},
	}
}

func Longitude() *Tree {
	return &Tree{
		Name: "Longitude",
		compare: func(r *Request) bool {
			last := strings.ToLower(r.LastName())
			return last == "longitude" && (r.Schema.IsNumber() || r.Schema.IsAny())
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.Longitude(), nil
		},
	}
}

func Latitude() *Tree {
	return &Tree{
		Name: "Latitude",
		compare: func(r *Request) bool {
			last := strings.ToLower(r.LastName())
			return last == "latitude" && (r.Schema.IsNumber() || r.Schema.IsAny())
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.Latitude(), nil
		},
	}
}

func Street() *Tree {
	return &Tree{
		Name: "Street",
		compare: func(r *Request) bool {
			last := strings.ToLower(r.LastName())
			return last == "street" && r.Schema.IsString()
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.Street(), nil
		},
	}
}

func OpenAddress() *Tree {
	return &Tree{
		Name: "OpenAddress",
		nodes: []*Tree{
			{
				Name: "Line1",
				compare: func(r *Request) bool {
					return strings.ToLower(r.LastName()) == "line1"
				},
				resolve: func(r *Request) (interface{}, error) {
					return gofakeit.Name(), nil
				},
			},
			{
				Name: "Line2",
				compare: func(r *Request) bool {
					return strings.ToLower(r.LastName()) == "line2"
				},
				resolve: func(r *Request) (interface{}, error) {
					return gofakeit.Street(), nil
				},
			},
			{
				Name: "Line3",
				compare: func(r *Request) bool {
					return strings.ToLower(r.LastName()) == "line3"
				},
				resolve: func(r *Request) (interface{}, error) {
					return fmt.Sprintf("%v %v %v", gofakeit.City(), gofakeit.StateAbr(), gofakeit.Zip()), nil
				},
			},
		},
		compare: func(r *Request) bool {
			name := r.GetName(-2)
			return strings.ToLower(name) == "address" || strings.HasSuffix(name, "Address")
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
		codeN, _ := strconv.ParseInt(code, 10, 64)
		return codeN
	}
	return code
}
