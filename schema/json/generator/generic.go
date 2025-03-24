package generator

import (
	"github.com/brianvoe/gofakeit/v6"
)

func Generic() *Tree {
	return &Tree{
		Name: "Generic",
		Nodes: []*Tree{
			Null(),
			Const(),
			Enum(),
		},
	}
}

func Null() *Tree {
	return &Tree{
		Name: "Null",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			if !s.IsNullable() {
				return false
			}

			if len(s.Type) == 1 {
				return true
			}

			n := gofakeit.Float32Range(0, 1)
			return n < 0.05
		},
		Fake: func(r *Request) (interface{}, error) {
			return nil, nil
		},
	}
}

func Const() *Tree {
	return &Tree{
		Name: "Const",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s != nil && s.Const != nil
		},
		Fake: func(r *Request) (interface{}, error) {
			return r.LastSchema().Const, nil
		},
	}
}

func Enum() *Tree {
	return &Tree{
		Name: "Enum",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s != nil && s.Enum != nil
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			return s.Enum[gofakeit.Number(0, len(s.Enum)-1)], nil
		},
	}
}
