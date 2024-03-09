package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
	"strings"
)

func PersonTree() *Tree {
	return &Tree{
		Name: "Person",
		nodes: []*Tree{
			Person(),
			People(),
		},
	}
}

func Person() *Tree {
	return &Tree{
		Name: "Person",
		compare: func(r *Request) bool {
			return !r.Schema.IsArray()

		},
		nodes: []*Tree{
			PersonName(),
			PersonFirstName(),
			PersonLastName(),
			Gender(),
			Phone(),
			Username(),
			Contact(),
			CreditCard(),
			PersonAny(),
		},
	}
}

func PersonName() *Tree {
	return &Tree{
		Name: "PersonName",
		compare: func(r *Request) bool {
			return r.matchLast([]string{"person", "name"}, true)
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.Name(), nil
		},
	}
}

func PersonFirstName() *Tree {
	return &Tree{
		Name: "PersonFirstName",
		compare: func(r *Request) bool {
			return r.matchLast([]string{"person", "firstname"}, true)
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.FirstName(), nil
		},
	}
}

func PersonLastName() *Tree {
	return &Tree{
		Name: "PersonLastName",
		compare: func(r *Request) bool {
			return r.matchLast([]string{"person", "lastname"}, true)
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.LastName(), nil
		},
	}
}

func Gender() *Tree {
	return &Tree{
		Name: "PersonGender",
		compare: func(r *Request) bool {
			return r.matchLast([]string{"gender"}, true) ||
				r.matchLast([]string{"sex"}, true)
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.Gender(), nil
		},
	}
}

func Phone() *Tree {
	return &Tree{
		Name: "Phone",
		compare: func(r *Request) bool {
			return r.matchLast([]string{"phone"}, true)
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.Phone(), nil
		},
	}
}

func Contact() *Tree {
	return &Tree{
		Name: "PetCategory",
		compare: func(r *Request) bool {
			if !r.Schema.IsObject() && !r.Schema.IsAny() {
				return false
			}
			return r.matchLast([]string{"contact"}, true)
		},
		resolve: func(r *Request) (interface{}, error) {
			contact := gofakeit.Contact()
			return map[string]interface{}{
				"phone": contact.Phone,
				"email": contact.Email,
			}, nil
		},
	}
}

func PersonAny() *Tree {
	return &Tree{
		Name: "PersonAny",
		compare: func(r *Request) bool {
			return r.LastName() == "person" && r.Schema.IsAny()
		},
		resolve: func(r *Request) (interface{}, error) {
			return map[string]interface{}{
				"firstname": gofakeit.FirstName(),
				"lastname":  gofakeit.LastName(),
				"gender":    gofakeit.Gender(),
				"email":     gofakeit.Email(),
			}, nil
		},
	}
}

func People() *Tree {
	return &Tree{
		Name: "People",
		compare: func(r *Request) bool {
			last := strings.ToLower(r.LastName())
			return (last == "persons" || last == "users" || last == "people") &&
				(r.Schema.IsArray() || r.Schema.IsAny())
		},
		resolve: func(r *Request) (interface{}, error) {
			next := r.With(Name("person"))
			if r.Schema.IsAny() {
				next = next.With(Schema(&schema.Schema{Type: []string{"array"}}))
			}
			return r.g.tree.Resolve(next)
		},
	}
}

func Username() *Tree {
	return &Tree{
		Name: "Username",
		compare: func(r *Request) bool {
			return r.LastName() == "username"
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.Username(), nil
		},
	}
}

func CreditCard() *Tree {
	return &Tree{
		Name: "CreditCard",
		compare: func(r *Request) bool {
			return strings.ToLower(r.LastName()) == "creditcard" || strings.ToLower(r.GetName(-2)) == "creditcard"
		},
		nodes: []*Tree{
			{
				Name: "CreditCardNumber",
				compare: func(r *Request) bool {
					return r.LastName() == "number"
				},
				resolve: func(r *Request) (interface{}, error) {
					return gofakeit.CreditCardNumber(nil), nil
				},
			},
			{
				Name: "CreditCardType",
				compare: func(r *Request) bool {
					return r.LastName() == "type"
				},
				resolve: func(r *Request) (interface{}, error) {
					return gofakeit.CreditCardType(), nil
				},
			},
			{
				Name: "CreditCardCvv",
				compare: func(r *Request) bool {
					return r.LastName() == "cvv"
				},
				resolve: func(r *Request) (interface{}, error) {
					return gofakeit.CreditCardCvv(), nil
				},
			},
			{
				Name: "CreditCardExp",
				compare: func(r *Request) bool {
					return r.LastName() == "exp"
				},
				resolve: func(r *Request) (interface{}, error) {
					return gofakeit.CreditCardExp(), nil
				},
			},
			{
				Name: "CreditCard",
				compare: func(r *Request) bool {
					return r.Schema.IsString() || r.Schema.IsAny()
				},
				resolve: func(r *Request) (interface{}, error) {
					if r.Schema.IsString() {
						return gofakeit.CreditCardNumber(nil), nil
					}
					if r.Schema.IsAny() {
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
