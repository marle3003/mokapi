package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
)

type Tree struct {
	Name string

	nodes []*Tree

	compare func(r *Request) bool
	resolve func(r *Request) (interface{}, error)
}

func (t *Tree) Add(node *Tree) {
	t.nodes = append(t.nodes, node)
}

func (t *Tree) Resolve(r *Request) (interface{}, error) {
	sel := compare(t, r)

	if sel != nil {
		return sel.resolve(r)
	}

	return nil, fmt.Errorf("unsupported schema: %v", r.Schema)
}

func compare(t *Tree, r *Request) *Tree {
	for _, n := range t.nodes {
		if n.compare != nil {
			if n.compare(r) {
				if n.resolve == nil {
					sel := compare(n, r)
					if sel != nil {
						return sel
					}
				} else {
					return n
				}
			}
		} else if n.compare == nil {
			if sel := compare(n, r); sel != nil {
				return sel
			}
		}
	}
	return nil
}

func NewTree() *Tree {
	root := &Tree{
		nodes: []*Tree{
			Null(),
			Const(),
			Enum(),
			AnyOf(),
			AllOf(),
			OneOf(),
			PetTree(),
			{
				Name:  "Names",
				nodes: []*Tree{AddressTree()},
				compare: func(r *Request) bool {
					return len(r.Names) > 0
				},
			},
			Number(),
			StringTree(),
			Object(),
			Array(),
			Bool(),
			Any(),
		},
	}

	return root
}

func Any() *Tree {
	simpleTypes := []string{
		"string",
		"number",
		"integer",
		"boolean",
	}
	complexTypes := []string{
		"array",
		"object",
	}
	types := append(simpleTypes, complexTypes...)

	getRandomType := func(r *Request) string {
		candidates := types
		if r.context["any"] == true {
			candidates = simpleTypes
		}
		i := gofakeit.Number(0, len(candidates)-1)
		return candidates[i]
	}

	return &Tree{
		Name: "Any",
		compare: func(r *Request) bool {
			return r.Schema.IsAny()
		},
		resolve: func(r *Request) (interface{}, error) {
			t := getRandomType(r)
			r = r.With(Schema(&schema.Schema{
				Type: []string{t},
			}))
			if _, ok := r.context["any"]; !ok {
				r.context["any"] = true
				defer delete(r.context, "any")
			}
			return r.g.tree.Resolve(r)
		},
	}
}

func Null() *Tree {
	return &Tree{
		Name: "Null",
		compare: func(r *Request) bool {
			if !r.Schema.IsNullable() {
				return false
			}
			n := gofakeit.Float32Range(0, 1)
			return n < 0.05
		},
		resolve: func(r *Request) (interface{}, error) {
			return nil, nil
		},
	}
}
