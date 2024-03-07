package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
	"mokapi/sortedmap"
	"unicode"
)

func Object() *Tree {
	return &Tree{
		nodes: []*Tree{
			Dictionary(),
			{
				nodes: []*Tree{},
				Name:  "Object",
				compare: func(r *Request) bool {
					return r.Schema.IsObject() && !r.Schema.IsFreeFrom()
				},
				resolve: func(r *Request) (interface{}, error) {
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
		compare: func(r *Request) bool {
			return r.Schema.IsObject() && r.Schema.IsFreeFrom()
		},
		resolve: func(r *Request) (interface{}, error) {
			r.Schema.Properties = &schema.Schemas{LinkedHashMap: sortedmap.LinkedHashMap[string, *schema.Ref]{}}

			minProps := 1
			maxProps := 10
			if r.Schema.MinProperties != nil {
				minProps = *r.Schema.MinProperties
			}
			if r.Schema.MaxProperties != nil {
				maxProps = *r.Schema.MaxProperties
			}

			length := gofakeit.Number(minProps, maxProps)
			if length == 0 {
				return map[string]interface{}{}, nil
			}

			for i := 0; i < length; i++ {

				name := firstLetterToLower(gofakeit.Noun())
				r.Schema.Properties.Set(name, nil)
			}
			return r.g.tree.Resolve(r)
		},
	}
}

func Dictionary() *Tree {
	return &Tree{
		Name: "Dictionary",
		compare: func(r *Request) bool {
			return r.Schema.IsDictionary()
		},
		resolve: func(r *Request) (interface{}, error) {
			length := gofakeit.Number(1, 10)
			var value *schema.Schema
			if r.Schema.AdditionalProperties.Ref != nil {
				value = r.Schema.AdditionalProperties.Ref.Value
			}
			m := map[string]interface{}{}
			for i := 0; i < length; i++ {
				v, err := r.g.tree.Resolve(r.With(Schema(value)))
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
	// recursion guard. Currently, we use a fixed depth: 1
	numRequestsSameAsThisOne := 0
	for _, h := range r.history {
		if r.Schema == h {
			numRequestsSameAsThisOne++
		}
	}
	if numRequestsSameAsThisOne > 1 {
		return nil, nil
	}

	m := map[string]interface{}{}

	if r.Schema.Properties == nil {
		return m, nil
	}

	for it := r.Schema.Properties.Iter(); it.Next(); {
		var propSchema *schema.Schema
		if it.Value() != nil {
			propSchema = it.Value().Value
		}
		prop := r.With(Name(it.Key()), Schema(propSchema))
		v, err := r.g.tree.Resolve(prop)
		if err != nil {
			return nil, err
		}
		m[it.Key()] = v
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
