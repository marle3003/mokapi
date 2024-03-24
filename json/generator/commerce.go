package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func Commerce() *Tree {
	return &Tree{
		Name: "Commerce",
		Nodes: []*Tree{
			Currency(),
			CreditCard(),
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
