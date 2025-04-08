package generator

import "github.com/brianvoe/gofakeit/v6"

func companyNodes() []*Node {
	return []*Node{
		{
			Name: "company",
			Children: []*Node{
				{
					Name: "company",
					Fake: fakeCompany,
				},
			},
		},
	}
}

func fakeCompany(r *Request) (any, error) {
	if r.Schema.IsString() || r.Schema.IsAny() {
		return gofakeit.Company(), nil
	}
	return nil, NotSupported
}
