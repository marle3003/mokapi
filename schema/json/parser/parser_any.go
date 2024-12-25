package parser

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
)

func (p *Parser) ParseAny(s *schema.Schema, data interface{}, evaluated map[string]bool) (interface{}, error) {
	var result *sortedmap.LinkedHashMap[string, interface{}]
	p2 := *p
	p2.ValidateAdditionalProperties = true

	isFreeFormUsed := false
NextAny:
	for _, any := range s.AnyOf {
		if any == nil || any.Value == nil {
			return p.parse(data, nil)
		}

		if any.Value.IsFreeForm() {
			// first we read data without free-form to get best matching data types
			// after we processed all schemas we read with free-form
			copySchema := *any.Value
			any = &schema.Ref{Value: &copySchema}
			isFreeFormUsed = true
			any.Value.AdditionalProperties = schema.NewRef(false)
		}

		if !any.IsObject() {
			r, err := p2.parse(data, any)
			if err == nil {
				return r, nil
			}
			continue
		}
		eval := map[string]bool{}
		o, err := p2.parseObject(data, any.Value, eval)
		if err != nil {
			var list *PathErrors
			if errors.As(err, &list) {
				for _, e := range *list {
					var pathError *PathError
					if !errors.As(e, &pathError) || pathError.Path != "additionalProperties" {
						continue NextAny
					}
				}
			} else {
				continue
			}
		}
		if result == nil {
			result = o
		} else {
			for it := o.Iter(); it.Next(); {
				if _, found := result.Get(it.Key()); !found {
					result.Set(it.Key(), it.Value())
				}
			}
		}

		for k, v := range eval {
			evaluated[k] = v
		}
	}

	if isFreeFormUsed {
		// read data with free-form and add only missing values
		// free-form returns never an error

		if result == nil {
			result = sortedmap.NewLinkedHashMap()
		}

		obj, _ := p.parseObject(data, &schema.Schema{Type: schema.Types{"object"}}, evaluated)
		if obj != nil {
			for it := obj.Iter(); it.Next(); {
				if _, found := result.Get(it.Key()); !found {
					result.Set(it.Key(), it.Value())
				}
			}
		}
	}

	if result == nil {
		return nil, fmt.Errorf("parse %v failed, expected %v", data, s)
	}

	if p.ConvertToSortedMap {
		return result, nil
	}

	return result.ToMap(), nil
}
