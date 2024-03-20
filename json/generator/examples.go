package generator

import "github.com/brianvoe/gofakeit/v6"

func Examples() *Tree {
	return &Tree{
		Name: "Examples",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s != nil && s.Examples != nil && len(s.Examples) > 0
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			return s.Examples[gofakeit.Number(0, len(s.Examples)-1)], nil
		},
	}
}
