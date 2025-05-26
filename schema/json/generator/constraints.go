package generator

import (
	"github.com/brianvoe/gofakeit/v6"
)

func applyConstraints(r *Request) (fakeFunc, bool) {
	if r.Schema == nil {
		return nil, false
	}

	switch {
	case len(r.Schema.Enum) > 0:
		return func() (interface{}, error) {
			v := pickEnumValue(r)
			return v, nil
		}, true
	case r.Schema.Const != nil:
		return func() (any, error) {
			return *r.Schema.Const, nil
		}, true
	}

	return nil, false
}

func pickEnumValue(r *Request) interface{} {
	var v any
	if len(r.Path) > 0 {
		last := r.Path[len(r.Path)-1]
		defer func() {
			r.Context.Values[last] = v
		}()
	}

	v = r.Schema.Enum[gofakeit.Number(0, len(r.Schema.Enum)-1)]
	return v
}
