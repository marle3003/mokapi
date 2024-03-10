package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
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
			AnyAddress(),
		},
	}
	return root
}

func AnyAddress() *Tree {
	return &Tree{
		Name: "AnyAddress",
		compare: func(r *Request) bool {
			return r.LastName() == "address" && r.Schema.IsAny()
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
				(r.Schema.IsNumber() || r.Schema.IsInteger() || r.Schema.IsAny())
		},
		resolve: func(r *Request) (interface{}, error) {
			return newPostCode(r.g, r.Schema)
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

func newPostCode(g *generator, s *schema.Schema) (interface{}, error) {
	if s == nil || s.IsAny() {
		s = &schema.Schema{Type: []string{"integer"}}
	}

	min := float64(10000)
	max := float64(99999)
	if s.Minimum == nil {
		s.Minimum = &min
	}
	if s.Maximum == nil {
		s.Maximum = &max
	}
	return g.tree.Resolve(&Request{Schema: s})
}
