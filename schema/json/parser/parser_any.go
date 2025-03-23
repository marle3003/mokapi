package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
)

func (p *Parser) ParseAny(s *schema.Schema, data interface{}, evaluated map[string]bool) (interface{}, error) {
	if m, ok := data.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
		p2 := *p
		p2.ValidateAdditionalProperties = true
		return p2.parseAnyObject(m, s, evaluated)
	}

	var result interface{}
	var err error
	for _, one := range s.AnyOf {
		result, err = p.parse(data, one)
		if err == nil {
			return result, nil
		}
	}

	return nil, &ErrorDetail{
		Message: "does not match any schemas",
		Field:   "anyOf",
	}
}

func (p *Parser) parseAnyObject(m *sortedmap.LinkedHashMap[string, interface{}], s *schema.Schema, evaluated map[string]bool) (interface{}, error) {
	var result *sortedmap.LinkedHashMap[string, interface{}]
	p2 := *p
	p2.ValidateAdditionalProperties = true
	var err error
	index := 0
	var one *schema.Schema

	for index, one = range s.AnyOf {
		if one == nil {
			result = m
			continue
		}

		eval := map[string]bool{}
		var obj *sortedmap.LinkedHashMap[string, interface{}]
		obj, err = p2.parseObject(m, one, eval)
		if err != nil {
			continue
		}

		var v interface{}
		v, err = p.evaluateUnevaluatedProperties(obj, one, eval)
		if err != nil {
			continue
		}
		obj = v.(*sortedmap.LinkedHashMap[string, interface{}])

		if result == nil {
			result = obj
		} else if obj != nil {
			for it := obj.Iter(); it.Next(); {
				if _, found := eval[it.Key()]; found {
					result.Set(it.Key(), it.Value())
				} else if _, found = result.Get(it.Key()); !found {
					result.Set(it.Key(), it.Value())
				}
			}
		}

		for k, v := range eval {
			evaluated[k] = v
		}
	}

	if result == nil {
		return nil, wrapErrorDetail(err, &ErrorDetail{
			Message: "does not match any schemas",
			Field:   fmt.Sprintf("anyOf/%d", index),
		})
	}

	return result, nil
}
