package v2

import "github.com/brianvoe/gofakeit/v6"

func products() []*Node {
	return []*Node{
		{
			Name: "product",
			Children: []*Node{
				{
					Name: "name",
					Fake: fakeProductName,
				},
				{
					Name: "description",
					Fake: fakeProductDescription,
				},
				{
					Name: "category",
					Fake: fakeProductCategory,
				},
				{
					Name: "material",
					Fake: fakeProductMaterial,
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
