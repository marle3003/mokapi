package generator

import (
	"strings"

	"github.com/brianvoe/gofakeit/v7"
)

func colors() []*Node {
	return []*Node{
		{
			Name:       "color",
			Attributes: []string{"color"},
			Fake:       fakeColor,
			Children: []*Node{
				{
					Name:       "name",
					Attributes: []string{"name"},
					Fake: func(r *Request) (any, error) {
						return gofakeit.Color(), nil
					},
				},
			},
		},
	}
}

func fakeColor(r *Request) (any, error) {
	s := r.Schema
	if s == nil {
		return gofakeit.Color(), nil
	}
	if s.MaxLength != nil && *s.MaxLength == 7 {
		return gofakeit.HexColor(), nil
	}

	if len(s.Examples) > 0 {
		str, ok := s.Examples[0].Value.(string)
		if ok && strings.HasPrefix(str, "#") {
			return gofakeit.HexColor(), nil
		}
	}

	if s.IsString() || s.IsAny() {
		return gofakeit.Color(), nil
	}

	return nil, NotSupported
}
