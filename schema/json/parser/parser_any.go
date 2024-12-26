package parser

import (
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

	return nil, Errorf("anyOf", "does not match any schemas of 'anyOf'")
}

func (p *Parser) parseAnyObject(m *sortedmap.LinkedHashMap[string, interface{}], s *schema.Schema, evaluated map[string]bool) (interface{}, error) {
	var result *sortedmap.LinkedHashMap[string, interface{}]
	p2 := *p
	p2.ValidateAdditionalProperties = true

	for _, one := range s.AnyOf {
		if one == nil || one.Value == nil {
			result = m
			continue
		}

		eval := map[string]bool{}
		obj, err := p2.parseObject(m, one.Value, eval)
		if err != nil {
			continue
		}
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
		return nil, Errorf("anyOf", "does not match any schemas of 'anyOf'")
	}

	return result, nil
}
