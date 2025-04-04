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
				return petCategory[gofakeit.Number(0, len(petCategory)-1)], nil
			}
			if last.Schema.HasProperties() {
				m := map[string]interface{}{}
				for it := last.Schema.Properties.Iter(); it.Next(); {
					if it.Key() == "name" {
						m["name"] = petCategory[gofakeit.Number(0, len(petCategory)-1)]
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

var petCategory = []string{"dog", "cat", "rabbit", "guinea pig", "hamster", "ferret", "hedgehog", "parrot", "canary", "turtle", "goldfish"}
