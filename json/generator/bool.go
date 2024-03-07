package generator

import "github.com/brianvoe/gofakeit/v6"

func Bool() *Tree {
	return &Tree{
		Name: "Boolean",
		compare: func(r *Request) bool {
			return r.Schema.Is("boolean")
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.Bool(), nil
		},
	}
}
