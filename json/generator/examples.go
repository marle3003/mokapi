package generator

import "github.com/brianvoe/gofakeit/v6"

func Examples() *Tree {
	return &Tree{
		Name: "Examples",
		Test: func(r *Request) bool {
			return r.Schema != nil && r.Schema.Examples != nil && len(r.Schema.Examples) > 0
		},
		Fake: func(r *Request) (interface{}, error) {
			return r.Schema.Examples[gofakeit.Number(0, len(r.Schema.Examples)-1)], nil
		},
	}
}
