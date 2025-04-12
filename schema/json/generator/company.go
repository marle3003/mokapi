package generator

import "github.com/brianvoe/gofakeit/v6"

func companyNodes() []*Node {
	return []*Node{
		{
			Name: "company",
			Fake: fakeCompany,
			Children: []*Node{
				{
					Name: "name",
					Fake: fakeCompany,
				},
			},
		},
		{
			Name: "industry",
			Fake: fakeIndustry,
		},
		{
			Name: "organization",
			Fake: fakeCompany,
			Children: []*Node{
				{
					Name: "name",
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

func fakeIndustry(r *Request) (any, error) {
	index := gofakeit.Number(0, len(industry)-1)
	return industry[index], nil
}

var (
	industry = []string{
		"Information Technology", "Healthcare", "Finance", "Retail", "Manufacturing", "Education",
		"Transportation and Logistics", "Real Estate", "Telecommunications", "Energy and Utilities",
		"Hospitality", "Entertainment and Media", "Construction", "Agriculture", "Aerospace and Defense",
	}

	organization = []string{}
)
