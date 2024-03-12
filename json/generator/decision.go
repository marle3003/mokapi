package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/pkg/errors"
	"mokapi/json/schema"
)

type Tree struct {
	Name string

	nodes []*Tree

	Test func(r *Request) bool
	Fake func(r *Request) (interface{}, error)
}

func (t *Tree) Add(node *Tree) {
	t.nodes = append(t.nodes, node)
}

func (t *Tree) Resolve(r *Request) (interface{}, error) {
	v, err := resolve(t, r)
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			return nil, fmt.Errorf("unsupported schema: %v", r.Schema)
		}
		return nil, err
	}
	return v, nil
}

func resolve(node *Tree, r *Request) (v interface{}, err error) {
	for _, n := range node.nodes {
		if n.Test != nil {
			if n.Test(r) {
				if n.Fake == nil {
					v, err = resolve(n, r)
				} else {
					v, err = n.Fake(r)
				}
				if err == nil || !errors.Is(err, ErrUnsupported) {
					return
				}
			}
		} else {
			v, err = resolve(n, r)
			if err == nil || !errors.Is(err, ErrUnsupported) {
				return
			}
		}
	}
	return nil, ErrUnsupported
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
			PersonTree(),
			AddressTree(),
			ProductTree(),
			NameTree(),
			Examples(),
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
		Test: func(r *Request) bool {
			return r.Schema.IsAny()
		},
		Fake: func(r *Request) (interface{}, error) {
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
		Test: func(r *Request) bool {
			if !r.Schema.IsNullable() {
				return false
			}
			n := gofakeit.Float32Range(0, 1)
			return n < 0.05
		},
		Fake: func(r *Request) (interface{}, error) {
			return nil, nil
		},
	}
}
