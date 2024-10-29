package parser

import (
	"errors"
	"fmt"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"reflect"
)

func (p *Parser) ParseAll(s *schema.Schema, data interface{}) (interface{}, error) {
	if len(s.AllOf) == 1 {
		return p.Parse(data, s.AllOf[0])
	}
	types, err := getTypeIntersection(s.AllOf)
	if err != nil {
		return nil, fmt.Errorf("allOf contains error: %w", err)
	}
	if len(types) == 0 {
		return nil, fmt.Errorf("allOf contains different types: %v", s)
	}

	switch reflect.ValueOf(data).Kind() {
	case reflect.Struct:
	case reflect.Map:
		if !types.Includes("object") {
			return nil, fmt.Errorf("parse value failed, got %v expected %v", toString(data), s)
		}
		return p.parseAllObject(s, data)
	default:
		for _, all := range s.AllOf {
			if all == nil || all.Value == nil {
				return nil, fmt.Errorf("schema is not defined")
			}

			copySchema := *all
			copySchema.Value.Type = types
			var err error
			data, err = p.Parse(data, &copySchema)
			if err != nil {
				return nil, err
			}
		}
	}

	return data, nil
}

func (p *Parser) parseAllObject(s *schema.Schema, data interface{}) (interface{}, error) {
	r := sortedmap.NewLinkedHashMap()

	isFreeFormUsed := false
	for _, all := range s.AllOf {

		if all == nil || all.Value == nil {
			return nil, fmt.Errorf("schema is not defined")
		}

		if all.Value.IsFreeForm() {
			// first we read data without free-form to get best matching data types
			// after we processed all schemas we read with free-form
			copySchema := *all.Value
			all = &schema.Ref{Value: &copySchema}
			isFreeFormUsed = true
			all.Value.AdditionalProperties = schema.AdditionalProperties{Forbidden: true}
		}

		obj, err := p.ParseObject(data, all.Value)
		if err != nil {
			var additionalError *AdditionalPropertiesNotAllowed
			if errors.As(err, &additionalError) {
				if additionalError.Schema != all.Value {
					return nil, fmt.Errorf("parse %v failed: value does not match part of allOf: %w", toString(data), err)
				}
				if uw, ok := err.(interface{ Unwrap() []error }); ok {
					errs := uw.Unwrap()
					if len(errs) > 1 {
						return nil, fmt.Errorf("parse %v failed: value does not match part of allOf: %w", toString(data), errors.Join(errs[:len(errs)-1]...))
					}
				}
			} else {
				return nil, fmt.Errorf("parse %v failed: value does not match part of allOf: %w", toString(data), err)
			}
		}

		for it := obj.Iter(); it.Next(); {
			if _, found := r.Get(it.Key()); !found {
				r.Set(it.Key(), it.Value())
			}
		}
	}

	if isFreeFormUsed {
		// read data with free-form and add only missing values
		// free-form returns never an error
		obj, _ := p.ParseObject(data, &schema.Schema{Type: schema.Types{"object"}})
		for it := obj.Iter(); it.Next(); {
			if _, found := r.Get(it.Key()); !found {
				r.Set(it.Key(), it.Value())
			}
		}
	}

	if p.ConvertToSortedMap {
		return r, nil
	}

	return r.ToMap(), nil
}

func getTypeIntersection(sets []*schema.Ref) (schema.Types, error) {
	m := map[string]struct{}{}

	countNoSchemaDefined := 0
	for _, set := range sets {
		if set == nil || set.Value == nil {
			countNoSchemaDefined++
			continue
		}
		if set.Value.Type == nil {
			// JSON schema does not require a type
			continue
		}

		if len(m) == 0 {
			for _, t := range set.Value.Type {
				m[t] = struct{}{}
			}
		} else {
			for k := range m {
				if !set.Value.Type.Includes(k) {
					delete(m, k)
				}
			}
		}
	}

	if len(sets) == countNoSchemaDefined {
		return nil, fmt.Errorf("no schema available")
	}

	var result schema.Types
	for k := range m {
		result = append(result, k)
	}
	return result, nil
}
