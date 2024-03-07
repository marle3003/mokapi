package generator

import "github.com/brianvoe/gofakeit/v6"

func Enum() *Tree {
	return &Tree{
		Name: "Enum",
		compare: func(r *Request) bool {
			return r.Schema != nil && r.Schema.Enum != nil
		},
		resolve: func(r *Request) (interface{}, error) {
			return r.Schema.Enum[gofakeit.Number(0, len(r.Schema.Enum)-1)], nil
		},
	}
}
