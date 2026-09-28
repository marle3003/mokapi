package generator

import "github.com/brianvoe/gofakeit/v7"

func products() []*Node {
	return []*Node{
		{
			Name:       "product",
			Attributes: []string{"product"},
			Children: []*Node{
				{
					Name:       "name",
					Attributes: []string{"name"},
					Fake:       fakeProductName,
				},
				{
					Name:       "description",
					Attributes: []string{"description"},
					Fake:       fakeProductDescription,
				},
				{
					Name:       "category",
					Attributes: []string{"category"},
					Fake:       fakeProductCategory,
				},
				{
					Name:       "material",
					Attributes: []string{"material"},
					Fake:       fakeProductMaterial,
				},
			},
		},
	}
}

func fakeProductName(r *Request) (any, error) {
	return gofakeit.ProductName(), nil
}

func fakeProductDescription(r *Request) (any, error) {
	return gofakeit.ProductDescription(), nil
}

func fakeProductCategory(r *Request) (any, error) {
	return gofakeit.ProductCategory(), nil
}

func fakeProductMaterial(r *Request) (any, error) {
	return gofakeit.ProductMaterial(), nil
}
