package generator

import "github.com/brianvoe/gofakeit/v6"

func Bool() *Tree {
	return &Tree{
		Name: "Boolean",
		Test: func(r *Request) bool {
			return r.Schema.Is("boolean")
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Bool(), nil
		},
	}
}
