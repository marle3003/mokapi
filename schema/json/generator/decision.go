package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/pkg/errors"
	"mokapi/schema/json/schema"
)

var (
	ErrUnsupported = errors.New("unsupported operation")
)

type Tree struct {
	Name   string `json:"name"`
	Custom bool   `json:"custom"`

	Test  func(r *Request) bool                 `json:"-"`
	Nodes []*Tree                               `json:"nodes,omitempty"`
	Fake  func(r *Request) (interface{}, error) `json:"-"`
}

func (t *Tree) Add(node *Tree) {
	t.Nodes = append(t.Nodes, node)
}

func (t *Tree) Resolve(r *Request) (interface{}, error) {
	v, err := resolve(t, r)
	if err != nil {

		if errors.Is(err, ErrUnsupported) {
			return nil, fmt.Errorf("unsupported schema: %v", r.Last().Schema)
		}
		return nil, err
	}
	return v, nil
}

func resolve(node *Tree, r *Request) (v interface{}, err error) {
	for _, n := range node.Nodes {
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
		Name: "Faker",
		Nodes: []*Tree{
			Generic(),
			Compositions(),
			Pets(),
			Personal(),
			Location(),
			Products(),
			It(),
			//Material(),
			Category(),
			Business(),
			Examples(),
			Basic(),
		},
	}

	return root
}

func Basic() *Tree {
	return &Tree{
		Name: "Basic",
		Nodes: []*Tree{
			Color(),
			Numbers(),
			Strings(),
			Object(),
			Array(),
			Bool(),
			AnyType(),
		},
	}
}

func AnyType() *Tree {
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
			return r.Path.MatchLast(IsSchemaAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			t := getRandomType(r)
			r = r.With(PathElements(&PathElement{
				Schema: &schema.Ref{
					Value: &schema.Schema{
						Type: []string{t},
					},
				},
			}))
			if _, ok := r.context["any"]; !ok {
				r.context["any"] = true
				defer delete(r.context, "any")
			}
			return r.g.tree.Resolve(r)
		},
	}
}

type RecursionGuard struct {
	s *schema.Ref
}

func (e *RecursionGuard) Error() string {
	return fmt.Sprintf("recursion in object path found but schema does not allow null: %v", e.s)
}
