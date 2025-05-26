package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/brianvoe/gofakeit/v6/data"
	"mokapi/schema/json/schema"
)

type currency struct {
	code string
	name string
}

var currencies map[string]currency

func financials() []*Node {
	return []*Node{
		{
			Name: "currency",
			Fake: fakeCurrency,
			Children: []*Node{
				{
					Name: "code",
					Fake: fakeCurrencyCode,
				},
				{
					Name:      "name",
					DependsOn: []string{"code"},
					Fake:      fakeCurrencyName,
				},
			},
		},
		{
			Name: "price",
			Fake: fakePriceValue,
			Children: []*Node{
				{
					Name: "value",
					Fake: fakePriceValue,
				},
				{
					Name: "amount",
					Fake: fakePriceValue,
				},
			},
		},
	}
}

func fakeCurrency(r *Request) (any, error) {
	if r.Schema.IsObject() {
		code, _ := fakeCurrencyCode(r)
		name, _ := fakeCurrencyName(r)
		return map[string]any{
			"code": code,
			"name": name,
		}, nil
	}
	return fakeCurrencyCode(r)
}

func fakeCurrencyCode(r *Request) (any, error) {
	v := gofakeit.CurrencyShort()
	r.Context.Values["currency"] = v
	return v, nil
}

func fakeCurrencyName(r *Request) (any, error) {
	if code, ok := r.Context.Values["currency"]; ok {
		ensureCurrencies()
		return currencies[code.(string)].name, nil
	}
	return gofakeit.Currency(), nil
}

func ensureCurrencies() {
	if currencies == nil {
		currencies = map[string]currency{}
		long := data.Currency["long"]
		for i, v := range data.Currency["short"] {
			c := currency{
				code: v,
				name: long[i],
			}
			currencies[c.code] = c
		}
	}
}

func fakePriceValue(r *Request) (any, error) {
	s := r.Schema
	if s.IsAny() {
		s = &schema.Schema{Type: []string{"number"}}
	}
	if s.IsInteger() {
		return fakeIntegerWithRange(s, 0, 1000000)
	}
	min, max := getRangeWithDefault(0, 1000000, s)
	return gofakeit.Price(min, max), nil
}
