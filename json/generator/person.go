package generator

import (
	"fmt"
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
		Test: func(r *Request) bool {
			return !r.Schema.IsArray()

		},
		nodes: []*Tree{
			PersonName(),
			FirstName(),
			LastName(),
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
		Test: func(r *Request) bool {
			return r.matchLast([]string{"person", "name"}, true)
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Name(), nil
		},
	}
}

func FirstName() *Tree {
	return &Tree{
		Name: "FirstName",
		Test: func(r *Request) bool {
			return strings.ToLower(r.LastName()) == "firstname"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.FirstName(), nil
		},
	}
}

func LastName() *Tree {
	return &Tree{
		Name: "LastName",
		Test: func(r *Request) bool {
			return strings.ToLower(r.LastName()) == "lastname"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.LastName(), nil
		},
	}
}

func Gender() *Tree {
	return &Tree{
		Name: "PersonGender",
		Test: func(r *Request) bool {
			return r.matchLast([]string{"gender"}, true) ||
				r.matchLast([]string{"sex"}, true)
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Gender(), nil
		},
	}
}

func Phone() *Tree {
	return &Tree{
		Name: "Phone",
		Test: func(r *Request) bool {
			if !r.Schema.IsString() && !r.Schema.IsAny() {
				return false
			}
			last := strings.ToLower(r.LastName())
			return last == "phone" || last == "phonenumber" || strings.HasSuffix(r.LastName(), "Phone")
		},
		Fake: func(r *Request) (interface{}, error) {
			countryCode := gofakeit.IntRange(1, 999)
			countryCodeLen := len(fmt.Sprintf("%v", countryCode))
			max := 15 - countryCodeLen
			min := 4
			if r.Schema != nil && r.Schema.MinLength != nil {
				min = *r.Schema.MinLength - countryCodeLen - 1
			}
			if r.Schema != nil && r.Schema.MaxLength != nil {
				max = *r.Schema.MaxLength - countryCodeLen - 1
			}
			nationalCodeLen := gofakeit.IntRange(min, max)
			return fmt.Sprintf("+%v%v", countryCode, gofakeit.Numerify(strings.Repeat("#", nationalCodeLen))), nil
		},
	}
}

func Contact() *Tree {
	return &Tree{
		Name: "Contact",
		Test: func(r *Request) bool {
			if !r.Schema.IsObject() && !r.Schema.IsAny() {
				return false
			}
			return r.matchLast([]string{"contact"}, true)
		},
		Fake: func(r *Request) (interface{}, error) {
			phone, err := r.g.tree.Resolve(r.With(Name("phone")))
			if err != nil {
				return nil, err
			}
			email, err := r.g.tree.Resolve(r.With(Name("email")))
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"phone": phone,
				"email": email,
			}, nil
		},
	}
}

func PersonAny() *Tree {
	return &Tree{
		Name: "PersonAny",
		Test: func(r *Request) bool {
			return r.LastName() == "person" && r.Schema.IsAny()
		},
		Fake: func(r *Request) (interface{}, error) {
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
		Test: func(r *Request) bool {
			last := strings.ToLower(r.LastName())
			return (last == "persons" || last == "users" || last == "people") &&
				(r.Schema.IsArray() || r.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
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
		Test: func(r *Request) bool {
			if !r.Schema.IsString() {
				return false
			}
			last := r.LastName()
			return strings.ToLower(last) == "username" || strings.HasSuffix(last, "UserName") || strings.HasSuffix(last, "Username")
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Username(), nil
		},
	}
}

func CreditCard() *Tree {
	isCreditCardObject := func(r *Request) bool {
		return strings.ToLower(r.LastName()) == "creditcard" || strings.ToLower(r.GetName(-2)) == "creditcard"
	}
	return &Tree{
		Name: "CreditCard",
		nodes: []*Tree{
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
					return isCreditCardObject(r) && (r.Schema.IsString() || r.Schema.IsAny())
				},
				Fake: func(r *Request) (interface{}, error) {
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
