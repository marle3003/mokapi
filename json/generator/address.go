package generator

import (
	"github.com/brianvoe/gofakeit/v6"
)

func AddressTree() *Tree {
	root := &Tree{
		Name: "Address",
		nodes: []*Tree{
			City(),
			Cities(),
			Country(),
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
