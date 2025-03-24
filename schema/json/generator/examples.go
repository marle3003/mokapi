package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	log "github.com/sirupsen/logrus"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"strings"
)

func Examples() *Tree {
	p := parser.Parser{}

	validate := func(v interface{}, s *schema.Schema) error {
		_, err := p.ParseWith(v, s)
		return err
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

			var path []string
			for _, item := range r.Path {
				path = append(path, item.Name)
			}

			// loop until valid example found
			for i := index; i < len(s.Examples); {
				v := s.Examples[i].Value
				if err := validate(v, s); err == nil {
					return v, nil
				} else {
					log.Warnf("skip using example from schema #/%s: example %d is not valid: %s", strings.Join(path, "/"), i, err.Error())
				}
				i = (i + 1) % len(s.Examples)
				if i == index {
					break
				}
			}

			log.Warnf("no example is valid: #/%v: generating a random value", strings.Join(path, "/"))
			return nil, ErrUnsupported
		},
	}
}
