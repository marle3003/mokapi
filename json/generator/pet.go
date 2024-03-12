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
		Test: func(r *Request) bool {
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
		Test: func(r *Request) bool {
			if !r.Schema.IsArray() && !r.Schema.IsAny() {
				return false
			}
			return strings.ToLower(r.GetName(-2)) == "pets"
		},
		Fake: func(r *Request) (interface{}, error) {
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
		Test: func(r *Request) bool {
			return r.matchLast([]string{"pet", "name"}, true)
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.PetName(), nil
		},
	}
}

func PetCategory() *Tree {
	return &Tree{
		Name: "PetCategory",
		Test: func(r *Request) bool {
			if r.Schema.IsArray() || r.Schema.IsObject() {
				return false
			}
			if r.matchLast([]string{"pet", "category"}, true) {
				return true
			}
			return r.matchLast([]string{"pet", "category", "name"}, true)
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Animal(), nil
		},
	}
}

func PetCategories() *Tree {
	return &Tree{
		Name: "PetCategories",
		Test: func(r *Request) bool {
			if !r.Schema.IsArray() && !r.Schema.IsAny() {
				return false
			}
			return r.matchLast([]string{"pet", "categories"}, true)
		},
		Fake: func(r *Request) (interface{}, error) {
			next := r.With(Name("pet", "category"))
			if r.Schema.IsAny() {
				next = next.With(Schema(&schema.Schema{Type: []string{"array"}}))
			}
			return r.g.tree.Resolve(next)
		},
	}
}
