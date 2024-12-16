package parser

import (
	"errors"
	"fmt"
	"mokapi/schema/json/schema"
	"reflect"
)

func (p *Parser) ParseArray(data interface{}, s *schema.Schema) (interface{}, error) {
	var result []interface{}

	switch list := data.(type) {
	case []interface{}:
		result = make([]interface{}, len(list))
		for i, e := range list {
			var err error
			result[i], err = p.Parse(e, s.Items)
			if err != nil {
				return nil, err
			}
		}
	case map[interface{}]interface{}:
		result = make([]interface{}, len(list))
		for _, v := range list {
			if i, err := p.Parse(v, s.Items); err != nil {
				return nil, err
			} else {
				result = append(result, i)
			}
		}
	default:
		v := reflect.ValueOf(data)
		if v == reflect.Zero(v.Type()) && !s.IsNullable() {
			if s.Default != nil {
				v = reflect.ValueOf(s.Default)
			} else {
				return nil, fmt.Errorf("TEST")
			}
		}
		if v.Kind() != reflect.Slice {
			return nil, fmt.Errorf("expected array but got: %v", toString(data))
		}
		var err error
		result = make([]interface{}, 0)
		for i := 0; i < v.Len(); i++ {
			o := v.Index(i)
			v, errItem := p.Parse(o.Interface(), s.Items)
			if errItem != nil {
				err = errors.Join(err, errItem)
			} else {
				result = append(result, v)
			}
		}
		if err != nil {
			return nil, err
		}
	}

	if err := validateArray(result, s); err != nil {
		return nil, err
	}

	return result, nil
}

func validateArray(a interface{}, s *schema.Schema) error {
	v := reflect.ValueOf(a)
	if s.MinItems != nil && v.Len() < *s.MinItems {
		return fmt.Errorf("should NOT have less than %v items, expected %v", *s.MinItems, s)
	}
	if s.MaxItems != nil && v.Len() > *s.MaxItems {
		return fmt.Errorf("should NOT have more than %v items, expected %v", *s.MaxItems, s)
	}

	if len(s.Enum) > 0 {
		return checkValueIsInEnum(a, s.Enum, &schema.Schema{Type: schema.Types{"array"}, Items: s.Items})
	}

	if s.UniqueItems {
		var unique []interface{}
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			for _, u := range unique {
				if compare(item, u) {
					return fmt.Errorf("should NOT have duplicate items (%v), expected %v", toString(item), s)
				}
			}
			unique = append(unique, item)
		}
	}

	return nil
}
