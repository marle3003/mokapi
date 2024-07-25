package parser

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
)

func (p *Parser) ParseAny(s *schema.Schema, data interface{}) (interface{}, error) {
	var result *sortedmap.LinkedHashMap[string, interface{}]

	isFreeFormUsed := false
	for _, any := range s.AnyOf {
		if any == nil || any.Value == nil {
			return p.Parse(data, nil)
		}

		if any.Value.IsFreeForm() {
			// first we read data without free-form to get best matching data types
			// after we processed all schemas we read with free-form
			copySchema := *any.Value
			any = &schema.Ref{Value: &copySchema}
			isFreeFormUsed = true
			any.Value.AdditionalProperties = schema.AdditionalProperties{Forbidden: true}
		}

		if !any.IsObject() {
			r, err := p.Parse(data, any)
			if err == nil {
				return r, nil
			}
			continue
		}
		o, err := p.ParseObject(data, any.Value)
		if err != nil {
			var additionalError *AdditionalPropertiesNotAllowed
			if errors.As(err, &additionalError) {
				if additionalError.Schema != any.Value {
					continue
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
	}

	if isFreeFormUsed {
		// read data with free-form and add only missing values
		// free-form returns never an error

		if result == nil {
			result = sortedmap.NewLinkedHashMap()
		}

		obj, _ := p.ParseObject(data, &schema.Schema{Type: schema.Types{"object"}})
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
