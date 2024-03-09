package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
	"strings"
)

func PetTree() *Tree {
	return &Tree{
		Name: "Pet",
		nodes: []*Tree{
			Pet(),
			Pets(),
			PetCategories(),
		},
	}
}

func Pet() *Tree {
	return &Tree{
		Name: "Pet",
		compare: func(r *Request) bool {
			return !r.Schema.IsArray()

		},
		nodes: []*Tree{
			PetName(),
			PetCategory(),
		},
	}
}

func Pets() *Tree {
	return &Tree{
		Name: "Pet",
		compare: func(r *Request) bool {
			if !r.Schema.IsArray() && !r.Schema.IsAny() {
				return false
			}
			return strings.ToLower(r.GetName(-2)) == "pets"
		},
		resolve: func(r *Request) (interface{}, error) {
			next := r.With(Name("pet", r.LastName()))
			if r.Schema.IsAny() {
				next = next.With(Schema(&schema.Schema{Type: []string{"array"}}))
			}
			return r.g.tree.Resolve(next)
		},
	}
}

func PetName() *Tree {
	return &Tree{
		Name: "PetName",
		compare: func(r *Request) bool {
			return r.matchLast([]string{"pet", "name"}, true)
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.PetName(), nil
		},
	}
}

func PetCategory() *Tree {
	return &Tree{
		Name: "PetCategory",
		compare: func(r *Request) bool {
			if r.Schema.IsArray() || r.Schema.IsObject() {
				return false
			}
			if r.matchLast([]string{"pet", "category"}, true) {
				return true
			}
			return r.matchLast([]string{"pet", "category", "name"}, true)
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.Animal(), nil
		},
	}
}

func PetCategories() *Tree {
	return &Tree{
		Name: "PetCategories",
		compare: func(r *Request) bool {
			if !r.Schema.IsArray() && !r.Schema.IsAny() {
				return false
			}
			return r.matchLast([]string{"pet", "categories"}, true)
		},
		resolve: func(r *Request) (interface{}, error) {
			next := r.With(Name("pet", "category"))
			if r.Schema.IsAny() {
				next = next.With(Schema(&schema.Schema{Type: []string{"array"}}))
			}
			return r.g.tree.Resolve(next)
		},
	}
}
