package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
)

func (p *Parser) ParseAll(s *schema.Schema, data interface{}, evaluated map[string]bool) (interface{}, error) {
	if m, ok := data.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
		p2 := *p
		p2.ValidateAdditionalProperties = true
		return p2.parseAllObject(m, s, evaluated)
	}

	var orig = data
	var err error
	for _, one := range s.AllOf {
		data, err = p.parse(data, one)
		if err != nil {
			return nil, fmt.Errorf("parse %v failed: does not match all schemas from 'allOf': %s: %w", orig, s, err)
		}
	}
	return data, nil
}

func (p *Parser) parseAllObject(m *sortedmap.LinkedHashMap[string, interface{}], s *schema.Schema, evaluated map[string]bool) (interface{}, error) {
	r := sortedmap.NewLinkedHashMap()
	err := &PathCompositionError{Path: "allOf", Message: "does not match all schemas from 'allOf'"}

	for index, one := range s.AllOf {
		path := fmt.Sprintf("%v", index)
		if one == nil {
			continue
		}

		if !one.IsObject() && !one.IsAny() {
			typeErr := Errorf("type", "invalid type, expected %v but got %v", one.Type.String(), toType(m))
			err.append(wrapError(path, typeErr))
			continue
		}

		eval := map[string]bool{}
		obj, oneErr := p.parseObject(m, one, eval)
		if oneErr != nil {
			err.append(wrapError(path, oneErr))
			continue
		}

		v, unEvalErr := p.evaluateUnevaluatedProperties(obj, one, eval)
		if unEvalErr != nil {
			err.append(wrapError(path, unEvalErr))
			continue
		}
		obj = v.(*sortedmap.LinkedHashMap[string, interface{}])

		if obj != nil {
			for it := obj.Iter(); it.Next(); {
				if _, found := eval[it.Key()]; found {
					r.Set(it.Key(), it.Value())
				} else if _, found = r.Get(it.Key()); !found {
					r.Set(it.Key(), it.Value())
				}
			}
		}
		for k, v := range eval {
			evaluated[k] = v
		}
	}

	if s.IsFreeForm() {
		for it := m.Iter(); it.Next(); {
			if _, found := r.Get(it.Key()); found {
				continue
			}
			r.Set(it.Key(), it.Value())
		}
	}

	if len(err.Errs) > 0 {
		return r, err
	}

	return r, nil
}
