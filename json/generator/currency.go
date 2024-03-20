package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
	"strings"
)

func Currency() *Tree {
	return &Tree{
		Name: "Currency",
		Nodes: []*Tree{
			ISO4217(),
			Price(),
			PriceObject(),
			Budget(),
		},
	}
}

func ISO4217() *Tree {
	return &Tree{
		Name: "ISO-4217",
		Test: func(r *Request) bool {
			last := r.Last()
			if strings.ToLower(last.Name) != "currency" {
				return false
			}
			if last.Schema.IsString() {
				if hasPattern(last.Schema) || hasFormat(last.Schema) {
					return false
				}
				if last.Schema.Value != nil && last.Schema.Value.MaxLength != nil && *last.Schema.Value.MaxLength > 3 {
					return false
				}
				return true
			}
			return last.Schema.IsAny()
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.CurrencyShort(), nil
		},
	}
}

func Price() *Tree {
	return &Tree{
		Name: "Price",
		Test: func(r *Request) bool {
			last := r.Last()
			return (last.Name == "price" || strings.HasSuffix(last.Name, "Price")) &&
				(last.Schema.IsAny() || last.Schema.IsInteger() || last.Schema.IsNumber())
		},
		Fake: func(r *Request) (interface{}, error) {
			return getPriceValue(r.LastSchema())
		},
	}
}

func PriceObject() *Tree {
	return &Tree{
		Name: "PriceObject",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(NameIgnoreCase("price"), Any())
		},
		Nodes: []*Tree{
			PriceValue(),
		},
	}
}

func PriceValue() *Tree {
	return &Tree{
		Name: "PriceValue",
		Test: func(r *Request) bool {
			last := r.Last()
			return last.Name == "value" && (last.Schema.IsAny() || last.Schema.IsInteger() || last.Schema.IsNumber())
		},
		Fake: func(r *Request) (interface{}, error) {
			return getPriceValue(r.LastSchema())
		},
	}
}

func Budget() *Tree {
	return &Tree{
		Name: "Budget",
		Test: func(r *Request) bool {
			last := r.Last()
			return (last.Name == "budget" || strings.HasSuffix(last.Name, "Budget")) &&
				(last.Schema.IsAny() || last.Schema.IsInteger())
		},
		Fake: func(r *Request) (interface{}, error) {
			return getPriceValue(r.LastSchema())
		},
	}
}

func getPriceValue(s *schema.Schema) (float64, error) {
	if s.IsAny() {
		s = &schema.Schema{Type: []string{"number"}}
	}
	min, max := getRangeWithDefault(s, 0, 1000000)
	return gofakeit.Price(min, max), nil
}
