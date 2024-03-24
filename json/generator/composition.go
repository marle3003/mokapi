package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
)

func Compositions() *Tree {
	return &Tree{
		Name: "Composition",
		Nodes: []*Tree{
			AnyOf(),
			AllOf(),
			OneOf(),
		},
	}
}

func AnyOf() *Tree {
	return &Tree{
		Name: "AnyOf",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s != nil && s.AnyOf != nil && len(s.AnyOf) > 0
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			i := gofakeit.Number(0, len(s.AnyOf)-1)
			return r.g.tree.Resolve(r.With(UsePathElement(r.LastName(), s.AnyOf[i])))
		},
	}
}

func AllOf() *Tree {
	return &Tree{
		Name: "AllOf",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s != nil && s.AllOf != nil && len(s.AllOf) > 0
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			result := map[string]interface{}{}
			for _, one := range s.AllOf {
				if one.IsAny() {
					if one.Value == nil {
						one.Value = &schema.Schema{Type: []string{"object"}}
					} else {
						one.Value.Type = []string{"object"}
					}
				}
				if !one.IsObject() {
					return nil, fmt.Errorf("allOf expects type of object but got %v", one.Type())
				}
				o, err := r.g.tree.Resolve(r.With(UsePathElement("", one)))
				if err != nil {
					return nil, fmt.Errorf("generate random data for schema failed: %v: %v", one, err)
				}
				m := o.(map[string]interface{})
				for k, v := range m {
					result[k] = v
				}
			}
			return result, nil
		},
	}
}

func OneOf() *Tree {
	return &Tree{
		Name: "OneOf",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s != nil && s.OneOf != nil && len(s.OneOf) > 0
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			selected := gofakeit.Number(0, len(s.OneOf)-1)
		Next:
			for t := 0; t < 20; t++ {
				v, err := r.g.tree.Resolve(r.With(UsePathElement("", s.OneOf[selected])))
				if err != nil {
					return nil, err
				}
				for i, one := range s.OneOf {
					if i == selected {
						continue
					}
					err = one.Validate(v)
					if err == nil {
						continue Next
					}
				}
				return v, nil
			}
			return nil, fmt.Errorf("to many tries to generate a random value for: %v", s)
		},
	}
}
