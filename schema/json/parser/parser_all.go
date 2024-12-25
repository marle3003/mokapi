package parser

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"reflect"
)

func (p *Parser) ParseAll(s *schema.Schema, data interface{}, evaluated map[string]bool) (interface{}, error) {
	switch reflect.ValueOf(data).Kind() {
	case reflect.Struct:
	case reflect.Map:
		return p.parseAllObject(s, data, evaluated)
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

func (p *Parser) parseAllObject(s *schema.Schema, data interface{}, evaluated map[string]bool) (interface{}, error) {
	r := sortedmap.NewLinkedHashMap()
	p2 := *p
	p2.ValidateAdditionalProperties = true
	err := &PathCompositionError{Path: "allOf", Message: "does not match all schemas from 'allOf'"}

	countNoSchemaDefined := 0
	isFreeFormUsed := false
	for index, one := range s.AllOf {
		path := fmt.Sprintf("%v", index)
		if one == nil || one.Value == nil {
			countNoSchemaDefined++
			continue
		}

		if !one.IsObject() && !one.IsAny() {
			typeErr := Errorf("type", "invalid type, expected %v but got %v", one.Type(), toType(data))
			err.append(wrapError(path, typeErr))
			continue
		}

		isLocalFreeForm := false
		if one.Value.IsFreeForm() {
			// first we read data without free-form to get best matching data types
			// after we processed all schemas we read with free-form
			copySchema := *one.Value
			one = &schema.Ref{Value: &copySchema}
			isFreeFormUsed = true
			isLocalFreeForm = true
			one.Value.AdditionalProperties = schema.NewRef(false)
		}

		eval := map[string]bool{}
		obj, oneErr := p2.parseObject(data, one.Value, eval)
		if oneErr != nil && isLocalFreeForm {
			var list *PathErrors
			if errors.As(oneErr, &list) {
				for _, e := range *list {
					var pathError *PathError
					if !errors.As(e, &pathError) {
						err.append(wrapError(path, e))
					} else if pathError.Path != "additionalProperties" {
						continue
					}
				}
			}
		} else if oneErr != nil {
			err.append(wrapError(path, oneErr))
			continue
		}

		if obj != nil {
			for it := obj.Iter(); it.Next(); {
				if _, found := r.Get(it.Key()); !found {
					r.Set(it.Key(), it.Value())
				} else if prop := one.Value.Properties.Get(it.Key()); !prop.IsAny() {
					// overwrite value with possible more precise type
					r.Set(it.Key(), it.Value())
				}
			}
		}
		for k, v := range eval {
			evaluated[k] = v
		}
	}

	if len(s.AllOf) == countNoSchemaDefined {
		return data, nil
	}

	if len(err.Errs) > 0 {
		return data, err
	}

	if isFreeFormUsed {
		// read data with free-form and add only missing values
		// free-form returns never an error
		obj, _ := p.parseObject(data, &schema.Schema{Type: schema.Types{"object"}}, evaluated)
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
