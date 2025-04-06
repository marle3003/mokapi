package v2

import "github.com/brianvoe/gofakeit/v6"

func pets() []*Node {
	return []*Node{
		{
			Name: "pet",
			Fake: fakePet,
			Children: []*Node{
				{
					Name: "name",
					Fake: fakePetName,
				},
				{
					Name: "category",
					Fake: fakePetCategory,
					Children: []*Node{
						{
							Name: "name",
							Fake: fakePetCategory,
						},
						{
							Name: "id",
							Fake: fakePetCategoryId,
						},
					},
				},
			},
		},
	}
}

func fakePet(r *Request) (any, error) {
	if r.Schema.IsObject() {
		name, _ := fakePetName(r)
		category, _ := fakeCategory(r)
		return map[string]any{
			"name":     name,
			"category": category,
		}, nil
	}
	return fakePetName(r)
}

func fakePetName(r *Request) (any, error) {
	if v, ok := r.ctx.store["pet.name"]; ok {
		return v, nil
	}

	v := gofakeit.PetName()
	r.ctx.store["pet.name"] = v
	return v, nil
}

func fakePetCategory(r *Request) (any, error) {
	if v, ok := r.ctx.store["pet.category"]; ok {
		return v, nil
	}

	index := gofakeit.Number(0, len(petCategory)-1)
	v := petCategory[index]
	r.ctx.store["pet.category"] = v
	return v, nil
}

func fakePetCategoryId(r *Request) (any, error) {
	if v, ok := r.ctx.store["pet.category.id"]; ok {
		return v, nil
	}

	v, err := fakeId(r)
	if err != nil {
		return nil, err
	}
	r.ctx.store["pet.category.id"] = v
	return v, nil
}

var petCategory = []string{"dog", "cat", "rabbit", "guinea pig", "hamster", "ferret", "hedgehog", "parrot", "canary", "turtle", "goldfish"}
