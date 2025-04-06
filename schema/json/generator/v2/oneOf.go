package v2

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/parser"
)

const maxOneOfTries = 20

func (r *resolver) oneOf(req *Request) (*faker, error) {
	s := req.Schema
	p := parser.Parser{}
	f := func() (any, error) {
		index := gofakeit.Number(0, len(s.OneOf)-1)
		selected := s.OneOf[index]
	Next:
		for t := 0; t < maxOneOfTries; t++ {
			fake, err := r.resolve(req.WithSchema(selected), true)
			if err != nil {
				return nil, err
			}
			v, err := fake.fake()
			if err != nil {
				return nil, err
			}
			for i, one := range s.OneOf {
				if i == index {
					continue
				}
				_, err = p.ParseWith(v, one)
				if err == nil {
					continue Next
				}
			}
			return v, nil
		}
		return nil, fmt.Errorf("to many tries to generate a random value for: %v", s)
	}
	return newFaker(f), nil
}
