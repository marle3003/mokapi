package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	log "github.com/sirupsen/logrus"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
)

func Examples() *Tree {
	p := parser.Parser{}

	validate := func(v interface{}, s *schema.Schema) error {
		_, err := p.ParseWith(v, &schema.Ref{Value: s})
		if err != nil {
			log.Warnf("skip using example from schema: %v: example is not valid: %v", s, parser.ToString(v))
			return ErrUnsupported
		}
		return nil
	}

	return &Tree{
		Name: "Examples",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s != nil && s.Examples != nil && len(s.Examples) > 0
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			// select random index
			index := gofakeit.Number(0, len(s.Examples)-1)

			// loop until valid example found
			for i := index; i < len(s.Examples); {
				v := s.Examples[i]
				if err := validate(v, s); err == nil {
					return v, nil
				}
				i = (i + 1) % len(s.Examples)
				if i == index {
					break
				}
			}
			log.Warnf("no example is valid: %v: generating a random value", s)
			return nil, ErrUnsupported
		},
	}
}
