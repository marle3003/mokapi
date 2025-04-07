package generator

import "github.com/brianvoe/gofakeit/v6"

func applyConstraints(r *Request) (fakeFunc, bool) {
	if r.Schema == nil {
		return nil, false
	}

	switch {
	case len(r.Schema.Enum) > 0:
		return func() (interface{}, error) {
			return pickEnumValue(r), nil
		}, true
	}

	return nil, false
}

func pickEnumValue(r *Request) interface{} {
	return r.Schema.Enum[gofakeit.Number(0, len(r.Schema.Enum)-1)]
}
