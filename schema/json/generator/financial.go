package generator

import (
	"fmt"
	"mokapi/schema/json/schema"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/brianvoe/gofakeit/v6/data"
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
		{
			Name:       "creditcard",
			Fake:       fakeCreditCard,
			Attributes: []string{"creditcard", "credit"},
			Children: []*Node{
				{
					Name: "card",
					Fake: fakeCreditCard,
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

func fakeCreditCard(r *Request) (any, error) {
	s := r.Schema

	if s != nil && len(s.Pattern) > 0 {
		return nil, NotSupported
	}

	minLength := 15
	maxLength := 19
	if s != nil {
		if s.MinLength != nil {
			minLength = *s.MinLength
		}
		if s.MaxLength != nil {
			maxLength = *s.MaxLength
		}
		if s.Minimum != nil {
			minLength = len(strconv.Itoa(int(*s.Minimum)))
		}
		if s.Maximum != nil {
			maxLength = len(strconv.Itoa(int(*s.Maximum)))
		}
	}
	n := gofakeit.Number(minLength, maxLength) - 1
	// major industry identifier
	mii := gofakeit.Number(1, 9)
	result := fmt.Sprintf("%d%s", mii, gofakeit.Numerify(strings.Repeat("#", n)))

	if s.IsString() {
		return result, nil
	}
	if s.IsInteger() || s.IsNumber() {
		return strconv.Atoi(result)
	}

	return nil, NotSupported
}
