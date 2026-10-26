package parser

import (
	"fmt"
	"mokapi/schema/json/schema"
	"reflect"
)

func (p *Parser) ParseArray(data any, s *schema.Schema, evaluated map[int]bool) (any, error) {
	if a, ok := data.([]any); ok {
		return p.parseArray(len(a), func(i int) any { return a[i] }, s, evaluated)
	}
	arr := reflect.ValueOf(data)
	if arr.Kind() == reflect.Pointer {
		arr = arr.Elem()
	}
	if arr.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected array but got: %v", ToString(data))
	}
	return p.parseArray(arr.Len(), func(i int) any { return arr.Index(i).Interface() }, s, evaluated)
}

func (p *Parser) parseArray(n int, at func(i int) any, s *schema.Schema, evaluated map[int]bool) (any, error) {
	var err ErrorList
	result := make([]any, 0, n)
	contains := 0
	for i := range n {
		errItem := ErrorDetail{}

		o := at(i)
		if s.PrefixItems != nil && i < len(s.PrefixItems) {
			v, errItems := p.parse(o, s.PrefixItems[i])
			if errItems != nil {
				errItem.append(errItems)
			} else {
				result = append(result, v)
			}
			evaluated[i] = true
		} else if s.Items != nil {
			v, errItems := p.parse(o, s.Items)
			if errItems != nil {
				errItem.append(errItems)
			} else {
				result = append(result, v)
			}
			evaluated[i] = true
		} else {
			result = append(result, o)
		}

		if s.Contains != nil {
			if _, errContains := p.parse(o, s.Contains); errContains == nil {
				contains++
			}
		}

		if len(errItem.Errors) > 0 {
			errItem.Field = fmt.Sprintf("items/%d", i)
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

	if s.UniqueItems != nil && *s.UniqueItems {
		var unique []any
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
