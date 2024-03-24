package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func Business() *Tree {
	return &Tree{
		Name: "Business",
		Nodes: []*Tree{
			Currency(),
			CreditCard(),
			DepartmentName(),
			CompanyName(),
			JobTitle(),
		},
	}
}

func CreditCard() *Tree {
	isCreditCardObject := func(r *Request) bool {
		name := strings.ToLower(r.LastName())
		return r.Path.MatchLast(NameIgnoreCase("creditcard"), Any()) || name == "creditcard"
	}
	return &Tree{
		Name: "CreditCard",
		Nodes: []*Tree{
			{
				Name: "CreditCardNumber",
				Test: func(r *Request) bool {
					return (isCreditCardObject(r) && r.LastName() == "number") || strings.ToLower(r.LastName()) == "creditcardnumber"
				},
				Fake: func(r *Request) (interface{}, error) {
					return gofakeit.CreditCardNumber(nil), nil
				},
			},
			{
				Name: "CreditCardType",
				Test: func(r *Request) bool {
					return (isCreditCardObject(r) && r.LastName() == "type") || strings.ToLower(r.LastName()) == "creditcardtype"
				},
				Fake: func(r *Request) (interface{}, error) {
					return gofakeit.CreditCardType(), nil
				},
			},
			{
				Name: "CreditCardCvv",
				Test: func(r *Request) bool {
					return r.LastName() == "cvv"
				},
				Fake: func(r *Request) (interface{}, error) {
					return gofakeit.CreditCardCvv(), nil
				},
			},
			{
				Name: "CreditCardExp",
				Test: func(r *Request) bool {
					return r.LastName() == "exp"
				},
				Fake: func(r *Request) (interface{}, error) {
					return gofakeit.CreditCardExp(), nil
				},
			},
			{
				Name: "CreditCard",
				Test: func(r *Request) bool {
					s := r.LastSchema()
					return isCreditCardObject(r) && (s.IsString() || s.IsAny())
				},
				Fake: func(r *Request) (interface{}, error) {
					s := r.LastSchema()
					if s.IsString() {
						return gofakeit.CreditCardNumber(nil), nil
					}
					if s.IsAny() {
						cc := gofakeit.CreditCard()
						return map[string]interface{}{
							"type":   cc.Type,
							"number": cc.Number,
							"cvv":    cc.Cvv,
							"exp":    cc.Exp,
						}, nil
					}
					return nil, ErrUnsupported
				},
			},
		},
	}
}

func DepartmentName() *Tree {
	return &Tree{
		Name: "DepartmentName",
		Test: func(r *Request) bool {
			last := r.Last()
			return strings.ToLower(last.Name) == "department" &&
				(last.Schema.IsAnyString() || last.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			index := gofakeit.Number(0, len(departments)-1)
			return departments[index], nil
		},
	}
}

func CompanyName() *Tree {
	return &Tree{
		Name: "ComanyName",
		Test: func(r *Request) bool {
			last := r.Last()
			return strings.ToLower(last.Name) == "company" &&
				(last.Schema.IsAnyString() || last.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			index := gofakeit.Number(0, len(companies)-1)
			return companies[index], nil
		},
	}
}

func JobTitle() *Tree {
	return &Tree{
		Name: "JobTitle",
		Test: func(r *Request) bool {
			last := r.Last()
			name := strings.ToLower(last.Name)
			return (name == "jobtitle" || name == "job_title") &&
				(last.Schema.IsAnyString() || last.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			index := gofakeit.Number(0, len(jobTitles)-1)
			return jobTitles[index], nil
		},
	}
}

var (
	departments = []string{"Marketing & Communications", "Human Resources", "Finance & Accounting", "Research & Development", "Sales & Business Development", "Customer Service & Support", "Information Technology", "Operations & Logistics", "Product Management", "Legal & Compliance", "Quality Assurance", "Administration & Facilities", "Supply Chain Management", "Public Relations", "Design & Creative", "Training & Development", "Procurement & Purchasing", "Health & Safety", "Sustainability & Environmental Affairs", "Internal Audit", "Corporate Strategy", "Business Intelligence", "Risk Management", "Project Management Office (PMO)", "Innovation & Technology Strategy", "Market Research", "Talent Acquisition", "Payroll & Benefits Administration", "Engineering & Development", "Client Success", "Content Creation", "Brand Management", "Regulatory Affairs", "Corporate Communications", "Event Planning & Management", "Distribution & Fulfillment", "Merchandising", "Training & Education", "Cybersecurity", "Strategic Partnerships", "Data Analytics", "Employee Relations", "Facilities Management", "Operations Planning", "Customer Experience", "International Business", "Logistics & Transportation", "Corporate Social Responsibility", "Financial Planning & Analysis", "Organizational Development"}
	companies   = []string{"Nexus Innovations", "Crestway Enterprises", "Horizon Dynamics", "Solstice Solutions", "Veritas Ventures", "Quantum Industries", "Stellar Strategies", "Catalyst Corporation", "Apex Alliance", "Synergy Systems", "Evolve Technologies", "Vanguard Ventures", "Phoenix Enterprises", "Zenith Innovations", "Summit Solutions", "Paradigm Partners", "Vertex Ventures", "Momentum Enterprises", "Fusion Dynamics", "Skyline Solutions", "Nexus Technologies", "Terra Nova Ventures", "Spectrum Solutions", "Echelon Enterprises", "Altitude Alliance"}
	jobTitles   = []string{"Senior Software Engineer", "Marketing Manager", "Human Resources Director", "Financial Analyst", "Product Designer", "Sales Representative", "Customer Success Manager", "Data Scientist", "Operations Coordinator", "Content Writer", "Project Manager", "Business Development Associate", "Quality Assurance Specialist", "Graphic Designer", "Legal Counsel", "Research Analyst", "Operations Manager", "Account Executive", "IT Support Specialist", "Social Media Manager", "Supply Chain Analyst", "UX/UI Designer", "Financial Controller", "HR Generalist", "Marketing Coordinator"}
)
