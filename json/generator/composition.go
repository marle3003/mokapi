package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
)

func AnyOf() *Tree {
	return &Tree{
		Name: "AnyOf",
		Test: func(r *Request) bool {
			return r.Schema != nil && r.Schema.AnyOf != nil && len(r.Schema.AnyOf) > 0
		},
		Fake: func(r *Request) (interface{}, error) {
			i := gofakeit.Number(0, len(r.Schema.AnyOf)-1)
			return r.g.tree.Resolve(r.With(Ref(r.Schema.AnyOf[i])))
		},
	}
}

func AllOf() *Tree {
	return &Tree{
		Name: "AllOf",
		Test: func(r *Request) bool {
			return r.Schema != nil && r.Schema.AllOf != nil && len(r.Schema.AllOf) > 0
		},
		Fake: func(r *Request) (interface{}, error) {
			result := map[string]interface{}{}
			for _, one := range r.Schema.AllOf {
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
				o, err := r.g.tree.Resolve(r.With(Ref(one)))
				if err != nil {
					return nil, fmt.Errorf("allOf expects to be valid against all of subschemas: %v", err)
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
			return r.Schema != nil && r.Schema.OneOf != nil && len(r.Schema.OneOf) > 0
		},
		Fake: func(r *Request) (interface{}, error) {
			selected := gofakeit.Number(0, len(r.Schema.OneOf)-1)
			v, err := r.g.tree.Resolve(r.With(Ref(r.Schema.OneOf[selected])))
			if err != nil {
				return nil, err
			}
			for i, one := range r.Schema.OneOf {
				if i == selected {
					continue
				}
				err = one.Validate(v)
				if err == nil {
					return nil, fmt.Errorf("value %v is valid to more as one schema", v)
				}
			}
			return v, nil
		},
	}
}
