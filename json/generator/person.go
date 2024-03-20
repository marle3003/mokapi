package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func Persons() *Tree {
	return &Tree{
		Name: "Person",
		Nodes: []*Tree{
			Person(),
			//People(),
			PersonAny(),
			Phone(),
			Contact(),
			CreditCard(),
		},
	}
}

func Person() *Tree {
	return &Tree{
		Name: "Person",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(NameIgnoreCase("person", "persons"), Any())

		},
		Nodes: []*Tree{
			PersonName(),
			FirstName(),
			LastName(),
			Gender(),
			Phone(),
			Username(),
		},
	}
}

func PersonName() *Tree {
	return &Tree{
		Name: "PersonName",
		Test: func(r *Request) bool {
			return r.LastName() == "name"
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
			name := r.LastName()
			return name == "gender" || name == "sex"
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
			last := r.Last()
			name := strings.ToLower(last.Name)
			return (name == "phone" || name == "phonenumber" || strings.HasSuffix(last.Name, "Phone")) &&
				(last.Schema.IsString() || last.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			countryCode := gofakeit.IntRange(1, 999)
			countryCodeLen := len(fmt.Sprintf("%v", countryCode))
			max := 15 - countryCodeLen
			min := 4
			s := r.LastSchema()
			if s != nil && s.MinLength != nil {
				min = *s.MinLength - countryCodeLen - 1
			}
			if s != nil && s.MaxLength != nil {
				max = *s.MaxLength - countryCodeLen - 1
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
			last := r.Last()
			if !last.Schema.IsObject() && !last.Schema.IsAny() {
				return false
			}
			return last.Name == "contact"
		},
		Fake: func(r *Request) (interface{}, error) {
			phone, err := r.g.tree.Resolve(r.With(UsePathElement("phone", nil)))
			if err != nil {
				return nil, err
			}
			email, err := r.g.tree.Resolve(r.With(UsePathElement("email", nil)))
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
		Name: "AnyPerson",
		Test: func(r *Request) bool {
			last := r.Last()
			return last.Name == "person" && last.Schema.IsAny()
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

//func People() *Tree {
//	return &Tree{
//		Name: "People",
//		Test: func(r *Request) bool {
//			last := strings.ToLower(r.LastName())
//			return (last == "persons" || last == "users" || last == "people") &&
//				(r.Schema.IsArray() || r.Schema.IsAny())
//		},
//		Fake: func(r *Request) (interface{}, error) {
//			next := r.With(Name("person"))
//			if r.Schema.IsAny() {
//				next = next.With(Schema(&schema.Schema{Type: []string{"array"}}))
//			}
//			return r.g.tree.Resolve(next)
//		},
//	}
//}

func Username() *Tree {
	return &Tree{
		Name: "Username",
		Test: func(r *Request) bool {
			last := r.Last()
			if !last.Schema.IsString() {
				return false
			}
			return strings.ToLower(last.Name) == "username" || strings.HasSuffix(last.Name, "UserName") ||
				strings.HasSuffix(last.Name, "Username")
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Username(), nil
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
