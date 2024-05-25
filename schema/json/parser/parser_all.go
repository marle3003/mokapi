package parser

import (
	"errors"
	"fmt"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
)

func (p *Parser) ParseAll(s *schema.Schema, data interface{}) (interface{}, error) {
	r := sortedmap.NewLinkedHashMap()

	isFreeFormUsed := false
	for _, all := range s.AllOf {
		if all == nil || all.Value == nil {
			return nil, fmt.Errorf("schema is not defined: allOf only supports type of object")
		}
		if !all.Value.Type.Includes("object") {
			return nil, fmt.Errorf("type of '%v' is not allowed: allOf only supports type of object", all.Value.Type.String())
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
