package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func Pets() *Tree {
	return &Tree{
		Name: "Pets",
		Nodes: []*Tree{
			//PetList(),
			PetCategory(),
			PetObject(),
			AnyPet(),
		},
	}
}

func PetObject() *Tree {
	return &Tree{
		Name: "PetObject",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(NameIgnoreCase("pets", "pet"), Any())
		},
		Nodes: []*Tree{
			PetName(),
			PetCategory(),
		},
	}
}

//func Pet() *Tree {
//	return &Tree{
//		Name: "Pet",
//		Test: func(r *Request) bool {
//			return !r.Schema.IsArray()
//		},
//		Nodes: []*Tree{
//			PetName(),
//			PetCategory(),
//		},
//	}
//}

func PetList() *Tree {
	return &Tree{
		Name: "PetList",
		Test: func(r *Request) bool {
			return r.LastSchema().IsArray()
		},
		Fake: func(r *Request) (interface{}, error) {
			r.Last().Name = "pet"
			return r.g.tree.Resolve(r)
		},
	}
}

func PetName() *Tree {
	return &Tree{
		Name: "PetName",
		Test: func(r *Request) bool {
			return r.LastName() == "name"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.PetName(), nil
		},
	}
}

func AnyPet() *Tree {
	return &Tree{
		Name: "Pet",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(ComparerFunc(func(p *PathElement) bool {
				return strings.ToLower(p.Name) == "pet" && (p.Schema.IsAny() || p.Schema.IsString() || (p.Schema.IsObject() && !p.Schema.HasProperties()))
			}))
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			if s.IsString() {
				return gofakeit.PetName(), nil
			}
			return nil, ErrUnsupported
		},
	}
}

func PetCategory() *Tree {
	return &Tree{
		Name: "PetCategory",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(NameIgnoreCase("pet"), ComparerFunc(func(p *PathElement) bool {
				return strings.ToLower(p.Name) == "category" && (p.Schema.IsString() || p.Schema.IsObject() || p.Schema.IsAny())
			}))

		},
		Fake: func(r *Request) (interface{}, error) {
			last := r.Last()
			if last.Schema.IsString() || last.Schema.IsAny() {
				return gofakeit.Animal(), nil
			}
			if last.Schema.HasProperties() {
				m := map[string]interface{}{}
				for it := last.Schema.Value.Properties.Iter(); it.Next(); {
					if it.Key() == "name" {
						m["name"] = gofakeit.Animal()
					} else {
						prop := r.With(UsePathElement(it.Key(), it.Value()))
						v, err := r.g.tree.Resolve(prop)
						if err != nil {
							return nil, err
						}
						m[it.Key()] = v
					}
				}
				return m, nil
			}
			return nil, ErrUnsupported
		},
	}
}

//
//func PetCategories() *Tree {
//	return &Tree{
//		Name: "PetCategories",
//		Test: func(r *Request) bool {
//			if !r.Schema.IsArray() && !r.Schema.IsAny() {
//				return false
//			}
//			return r.matchLast([]string{"pet", "categories"}, true)
//		},
//		Fake: func(r *Request) (interface{}, error) {
//			next := r.With(Name("pet", "category"))
//			if r.Schema.IsAny() {
//				next = next.With(Schema(&schema.Schema{Type: []string{"array"}}))
//			}
//			return r.g.tree.Resolve(next)
//		},
//	}
//}
