package parser

import (
	"errors"
	"fmt"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"reflect"
)

func (p *Parser) ParseAll(s *schema.Schema, data interface{}) (interface{}, error) {
	switch reflect.ValueOf(data).Kind() {
	case reflect.Struct:
	case reflect.Map:
		obj, err := p.parseAllObject(s, data)
		if err != nil {
			return nil, fmt.Errorf("parse %s failed: does not match %s:\n%w", ToString(data), s, err)
		}
		return obj, nil
	}

	var orig = data
	var err error
	for _, one := range s.AllOf {
		data, err = p.Parse(data, one)
		if err != nil {
			return nil, fmt.Errorf("parse %v failed: does not match %s: %w", orig, s, err)
		}
	}
	return data, nil
}

func (p *Parser) parseAllObject(s *schema.Schema, data interface{}) (interface{}, error) {
	r := sortedmap.NewLinkedHashMap()
	var err error

	countNoSchemaDefined := 0
	isFreeFormUsed := false
	for _, one := range s.AllOf {
		if one == nil || one.Value == nil {
			countNoSchemaDefined++
			continue
		}

		if !one.IsObject() {
			return nil, fmt.Errorf("invalid type object, expected %s", one.String())
		}

		isLocalFreeForm := false
		if one.Value.IsFreeForm() {
			// first we read data without free-form to get best matching data types
			// after we processed all schemas we read with free-form
			copySchema := *one.Value
			one = &schema.Ref{Value: &copySchema}
			isFreeFormUsed = true
			isLocalFreeForm = true
			one.Value.AdditionalProperties = schema.AdditionalProperties{Forbidden: true}
		}

		obj, oneErr := p.ParseObject(data, one.Value)
		if oneErr != nil {
			if isLocalFreeForm {
				var additionalError *AdditionalPropertiesNotAllowed
				if errors.As(oneErr, &additionalError) {
					if additionalError.Schema != one.Value {
						err = errors.Join(err, removeFreeForm(oneErr))
						continue
					}
					if uw, ok := oneErr.(interface{ Unwrap() []error }); ok {
						errs := uw.Unwrap()
						if len(errs) > 1 {
							err = errors.Join(err, removeFreeForm(errors.Join(errs[:len(errs)-1]...)))
							continue
						}
					}
				} else {
					err = errors.Join(err, removeFreeForm(oneErr))
					continue
				}
			} else {
				err = errors.Join(err, oneErr)
				continue
			}
		}

		for it := obj.Iter(); it.Next(); {
			if _, found := r.Get(it.Key()); !found {
				r.Set(it.Key(), it.Value())
			} else if prop := one.Value.Properties.Get(it.Key()); !prop.IsAny() {
				// overwrite value with possible more precise type
				r.Set(it.Key(), it.Value())
			}
		}
	}

	if len(s.AllOf) == countNoSchemaDefined {
		return data, nil
	}

	if err != nil {
		return nil, err
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

func removeFreeForm(errs ...error) error {
	var r []error
	for _, err := range errs {
		s := err.Error()
		n := len(s)
		s = s[0 : n-len(" free-form=false")]
		r = append(r, fmt.Errorf(s))
	}
	return errors.Join(r...)
}
