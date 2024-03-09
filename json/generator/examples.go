package generator

import "github.com/brianvoe/gofakeit/v6"

func Examples() *Tree {
	return &Tree{
		Name: "Examples",
		compare: func(r *Request) bool {
			return r.Schema != nil && r.Schema.Examples != nil && len(r.Schema.Examples) > 0
		},
		resolve: func(r *Request) (interface{}, error) {
			return r.Schema.Examples[gofakeit.Number(0, len(r.Schema.Examples)-1)], nil
		},
	}
}
