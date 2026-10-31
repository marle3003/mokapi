package parser

import (
	"fmt"
	"maps"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
)

func (p *Parser) ParseAll(s *schema.Schema, data any, evaluated map[string]bool) (any, error) {
	if m, ok := data.(*sortedmap.LinkedHashMap[string, any]); ok {
		return p.parseAllObject(m, s, evaluated)
	}

	allOfError := &ErrorComposition{
		Message: "does not match all schema",
		Field:   "allOf",
	}

	var err error
	for i, one := range s.AllOf {
		data, err = p.parse(data, one)
		if err != nil {
			allOfError.Errors = append(allOfError.Errors,
				wrapErrorDetail(
					err,
					&ErrorDetail{
						Field: fmt.Sprintf("%d", i),
					},
				),
			)
		}
	}

	if len(allOfError.Errors) > 0 {
		return nil, allOfError
	}

	return data, nil
}

func (p *Parser) parseAllObject(m *sortedmap.LinkedHashMap[string, any], s *schema.Schema, evaluated map[string]bool) (any, error) {
	r := sortedmap.NewLinkedHashMap()
	err := ErrorList{}

	for index, one := range s.AllOf {
		path := fmt.Sprintf("%v", index)
		if one == nil {
			continue
		}

		if !one.IsObject() && !one.IsAny() {
			typeErr := &ErrorDetail{
				Message: fmt.Sprintf("invalid type, expected %v but got %v", one.Type.String(), toType(m)),
				Field:   "type",
			}
			err = append(err, wrapErrorDetail(typeErr, &ErrorDetail{Field: path}))
			continue
		}

		eval := map[string]bool{}
		obj, oneErr := p.parseObjectWithComposition(m, one, eval)
		if oneErr != nil {
			err = append(err, wrapErrorDetail(oneErr, &ErrorDetail{Field: path}))
			continue
		}

		v, unEvalErr := p.evaluateUnevaluatedProperties(obj, one, eval)
		if unEvalErr != nil {
			err = append(err, wrapErrorDetail(unEvalErr, &ErrorDetail{Field: path}))
			continue
		}
		obj = v.(*sortedmap.LinkedHashMap[string, any])

		if obj != nil {
			for it := obj.Iter(); it.Next(); {
				if _, found := eval[it.Key()]; found {
					r.Set(it.Key(), it.Value())
				} else if _, found = r.Get(it.Key()); !found {
					r.Set(it.Key(), it.Value())
				}
			}
		}
		maps.Copy(evaluated, eval)
	}

	if s.IsFreeForm() {
		for it := m.Iter(); it.Next(); {
			if _, found := r.Get(it.Key()); found {
				continue
			}
			r.Set(it.Key(), it.Value())
		}
	}

	if len(err) > 0 {
		return r, &ErrorComposition{
			Message: "does not match all schema",
			Field:   "allOf",
			Errors:  err,
		}
	}

	return r, nil
}
