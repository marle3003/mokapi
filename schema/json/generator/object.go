package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"unicode"
)

func Object() *Tree {
	return &Tree{
		Name: "Objects",
		Nodes: []*Tree{
			Dictionary(),
			{
				Nodes: []*Tree{},
				Name:  "Object",
				Test: func(r *Request) bool {
					return r.Path.MatchLast(ComparerFunc(func(p *PathElement) bool {
						return p.Schema.IsObject() && p.Schema.HasProperties()
					}))
				},
				Fake: func(r *Request) (interface{}, error) {
					return createObject(r)
				},
			},
			AnyObject(),
		},
	}
}

func AnyObject() *Tree {
	return &Tree{
		Name: "AnyObject",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s.IsObject() && s.IsFreeForm()
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			s.Properties = &schema.Schemas{LinkedHashMap: sortedmap.LinkedHashMap[string, *schema.Schema]{}}

			minProps := 1
			maxProps := 10
			if s.MinProperties != nil {
				minProps = *s.MinProperties
			}
			if s.MaxProperties != nil {
				maxProps = *s.MaxProperties
			}

			length := gofakeit.Number(minProps, maxProps)
			if length == 0 {
				return map[string]interface{}{}, nil
			}

			for i := 0; i < length; i++ {

				name := firstLetterToLower(gofakeit.Noun())
				s.Properties.Set(name, nil)
			}
			return r.g.tree.Resolve(r)
		},
	}
}

func Dictionary() *Tree {
	return &Tree{
		Name: "Dictionary",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			if s == nil {
				return false
			}
			return s.AdditionalProperties != nil
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			length := gofakeit.Number(1, 10)
			m := map[string]interface{}{}
			for i := 0; i < length; i++ {
				v, err := r.g.tree.Resolve(r.With(UsePathElement("", s.AdditionalProperties)))
				if err != nil {
					return nil, err
				}
				key := firstLetterToLower(gofakeit.Noun())
				m[key] = v
			}
			return m, nil
		},
	}
}

func createObject(r *Request) (interface{}, error) {
	s := r.Last().Schema
	// recursion guard. Currently, we use a fixed depth: 1
	numRequestsSameAsThisOne := 0
	for _, h := range r.history {
		if s == h {
			numRequestsSameAsThisOne++
		}
	}
	if numRequestsSameAsThisOne > 1 {
		if !s.IsNullable() {
			return nil, &RecursionGuard{s: s}
		}
		return nil, nil
	}

	m := map[string]interface{}{}

	if s.Properties == nil {
		return m, nil
	}

	for it := s.Properties.Iter(); it.Next(); {
		prop := r.With(UsePathElement(it.Key(), it.Value()))
		v, err := r.g.tree.Resolve(prop)
		if err != nil {
			return nil, err
		}
		m[it.Key()] = v
	}

	if s.If != nil {
		p := parser.Parser{}
		_, err := p.ParseWith(m, s.If)
		var cond *schema.Schema
		if err == nil && s.Then != nil {
			cond = s.Then
		} else if err != nil && s.Else != nil {
			cond = s.Else
		}
		if cond != nil {
			v, err := r.g.tree.Resolve(r.With(UsePathElement(r.LastName(), cond)))
			if err != nil {
				return nil, err
			}
			if m2, ok := v.(map[string]interface{}); ok {
				for key, val := range m2 {
					m[key] = val
				}
			}
		}
	}

	return m, nil
}

func firstLetterToLower(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToLower(r[0])

	return string(r)
}
