package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
	"reflect"
)

func (p *Parser) ParseArray(data interface{}, s *schema.Schema, evaluated map[int]bool) (interface{}, error) {
	arr := reflect.ValueOf(data)
	if arr.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected array but got: %v", ToString(data))
	}
	var err ErrorList
	result := make([]interface{}, 0)
	contains := 0
	for i := 0; i < arr.Len(); i++ {
		errItem := ErrorDetail{Field: fmt.Sprintf("items/%d", i)}

		o := arr.Index(i)
		if s.PrefixItems != nil && i < len(s.PrefixItems) {
			v, errItems := p.parse(o.Interface(), s.PrefixItems[i])
			if errItems != nil {
				errItem.append(errItems)
			} else {
				result = append(result, v)
			}
			evaluated[i] = true
		} else if s.Items != nil {
			v, errItems := p.parse(o.Interface(), s.Items)
			if errItems != nil {
				errItem.append(errItems)
			} else {
				result = append(result, v)
			}
			evaluated[i] = true
		} else {
			result = append(result, o.Interface())
		}

		if s.Contains != nil {
			if _, errContains := p.parse(o.Interface(), s.Contains); errContains == nil {
				contains++
			}
		}

		if len(errItem.Errors) > 0 {
			err = append(err, &errItem)
		}
	}

	if s.Contains != nil {
		if s.MinContains != nil && contains < *s.MinContains {
			err = append(err, &ErrorDetail{
				Message: fmt.Sprintf("contains match count %v is less than minimum contains count of %v", contains, *s.MinContains),
				Field:   "minContains",
			})
		}
		if s.MaxContains != nil && contains > *s.MaxContains {
			err = append(err, &ErrorDetail{
				Message: fmt.Sprintf("contains match count %v exceeds maximum contains count of %v", contains, *s.MaxContains),
				Field:   "maxContains",
			})
		}
		if s.MinContains == nil && contains == 0 {
			err = append(err, &ErrorDetail{
				Message: "no items match contains",
				Field:   "contains",
			})
		}
	}

	if s.MinItems != nil && len(result) < *s.MinItems {
		err = append(err, &ErrorDetail{
			Message: fmt.Sprintf("item count %v is less than minimum count of %v", len(result), *s.MinItems),
			Field:   "minItems",
		})
	}
	if s.MaxItems != nil && len(result) > *s.MaxItems {
		err = append(err, &ErrorDetail{
			Message: fmt.Sprintf("item count %v exceeds maximum count of %v", len(result), *s.MaxItems),
			Field:   "maxItems",
		})
	}

	if len(s.Enum) > 0 {
		if errEnum := checkValueIsInEnum(result, s.Enum, &schema.Schema{Type: schema.Types{"array"}, Items: s.Items}); errEnum != nil {
			err = append(err, errEnum)
		}
	}

	if s.UniqueItems {
		var unique []interface{}
		for i, item := range result {
			for _, u := range unique {
				if compare(item, u) {
					err = append(err, &ErrorDetail{
						Message: fmt.Sprintf("non-unique array item at index %v", i),
						Field:   "uniqueItems",
					})
				}
			}
			unique = append(unique, item)
		}
	}

	if err != nil {
		return result, &err
	}

	return result, nil
}
