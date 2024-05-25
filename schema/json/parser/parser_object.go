package parser

import (
	"errors"
	"fmt"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"reflect"
	"sort"
	"strings"
)

type AdditionalPropertiesNotAllowed struct {
	Properties []string
	Schema     *schema.Schema
}

func (e *AdditionalPropertiesNotAllowed) Error() string {
	return fmt.Sprintf("additional properties '%v' not allowed, expected %v", strings.Join(e.Properties, ", "), e.Schema)
}

func (p *Parser) ParseObject(data interface{}, s *schema.Schema) (*sortedmap.LinkedHashMap[string, interface{}], error) {
	var result *sortedmap.LinkedHashMap[string, interface{}]
	var err error

	if m, ok := data.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
		result, err = p.parseLinkedMap(m, s)
	} else {

		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Struct:
			result, err = p.parseStruct(v, s)
		case reflect.Map:
			result, err = p.parseMap(v, s)
		default:
			return nil, fmt.Errorf("encode '%v' to %v failed", data, s)
		}
	}

	return result, err
}

func (p *Parser) parseLinkedMap(m *sortedmap.LinkedHashMap[string, interface{}], s *schema.Schema) (*sortedmap.LinkedHashMap[string, interface{}], error) {
	if s.IsDictionary() {
		return m, validateObject(m, s)
	}

	obj := sortedmap.NewLinkedHashMap()
	for it := m.Iter(); it.Next(); {
		name := it.Key()

		var field *schema.Ref
		if s.Properties != nil {
			field, _ = s.Properties.Get(name)
		}

		if field != nil || s.IsFreeForm() {
			d, err := p.Parse(it.Value(), field)
			if err != nil {
				return nil, err
			}
			obj.Set(name, d)
		}
	}

	err := validateObject(obj, s)

	return obj, err
}

func (p *Parser) parseStruct(v reflect.Value, s *schema.Schema) (*sortedmap.LinkedHashMap[string, interface{}], error) {
	t := v.Type()
	obj := sortedmap.NewLinkedHashMap()
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		name := unTitle(ft.Name)
		val := v.Field(i)

		if prop, ok := s.Properties.Get(name); ok || s.IsFreeForm() {
			d, err := p.Parse(val.Interface(), prop)
			if err != nil {
				return nil, fmt.Errorf("encode property '%v' failed: %w", name, err)
			}
			obj.Set(name, d)
		}
	}

	err := validateObject(obj, s)

	return obj, err
}

func (p *Parser) parseMap(v reflect.Value, s *schema.Schema) (*sortedmap.LinkedHashMap[string, interface{}], error) {
	obj := sortedmap.NewLinkedHashMap()
	var err error

	if s.HasProperties() {
		for it := s.Properties.Iter(); it.Next(); {
			name := it.Key()
			o := v.MapIndex(reflect.ValueOf(name))
			if !o.IsValid() {
				continue
			}
			d, valErr := p.Parse(o.Interface(), it.Value())
			if valErr != nil {
				err = errors.Join(err, fmt.Errorf("parse property '%v' failed: %w", name, valErr))
			}
			obj.Set(name, d)
		}
	}

	if s.IsDictionary() {
		for _, k := range v.MapKeys() {
			name := fmt.Sprintf("%v", k.Interface())
			if _, found := obj.Get(name); !found {
				o := v.MapIndex(k)
				d, valErr := p.Parse(o.Interface(), s.AdditionalProperties.Ref)
				if valErr != nil {
					err = errors.Join(err, valErr)
				}
				obj.Set(name, d)
			}
		}
	}

	if s.IsFreeForm() || s.IsDictionary() {
		for _, k := range v.MapKeys() {
			name := fmt.Sprintf("%v", k.Interface())
			if _, found := obj.Get(name); !found {
				o := v.MapIndex(k)
				obj.Set(name, o.Interface())
			}
		}
	}

	valErr := validateObject(obj, s)
	if valErr != nil {
		err = errors.Join(err, valErr)
	}

	if !s.IsFreeForm() && p.ValidateAdditionalProperties {
		var additionalProps []string
		for _, vKey := range v.MapKeys() {
			key := fmt.Sprintf("%v", vKey.Interface())
			_, ok := s.Properties.Get(key)
			if !ok {
				additionalProps = append(additionalProps, key)
			}
		}
		if len(additionalProps) > 0 {
			sort.Strings(additionalProps)
			err = errors.Join(err, &AdditionalPropertiesNotAllowed{Properties: additionalProps, Schema: s})
		}
	}

	return obj, err
}

func (p *Parser) getProperty(name string, props *schema.Schemas) (string, *schema.Ref) {
	prop, _ := props.Get(name)
	return name, prop
}

func validateObject(i interface{}, s *schema.Schema) error {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Map {
		if s.MinProperties != nil && v.Len() < *s.MinProperties {
			return fmt.Errorf("validation error minProperties on %v, expected %v", toString(i), s)
		}
		if s.MaxProperties != nil && v.Len() > *s.MaxProperties {
			return fmt.Errorf("validation error maxProperties on %v, expected %v", toString(i), s)
		}
		if !s.IsFreeForm() && s.Properties != nil {
			var add []string
			for _, k := range v.MapKeys() {
				name := k.Interface().(string)
				if _, ok := s.Properties.Get(name); !ok {
					add = append(add, name)
				}
			}
			if len(add) > 0 {
				sort.Strings(add)
				return &AdditionalPropertiesNotAllowed{Properties: add, Schema: s}
			}
		}

		for _, p := range s.Required {
			if e := v.MapIndex(reflect.ValueOf(p)); !e.IsValid() {
				return fmt.Errorf("missing required field '%v', expected %v", p, s)
			}
		}
	} else if m, ok := i.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
		if s.MinProperties != nil && m.Len() < *s.MinProperties {
			return fmt.Errorf("validation error minProperties on %v, expected %v", m, s)
		}
		if s.MaxProperties != nil && m.Len() > *s.MaxProperties {
			return fmt.Errorf("validation error maxProperties on %v, expected %v", m, s)
		}

		if !s.IsFreeForm() && s.Properties != nil {
			var add []string
			for it := m.Iter(); it.Next(); {
				name := it.Key()
				if _, ok := s.Properties.Get(name); !ok {
					add = append(add, name)
				}
			}
			if len(add) > 0 {
				sort.Strings(add)
				return &AdditionalPropertiesNotAllowed{Properties: add, Schema: s}
			}
		}

		for _, p := range s.Required {
			if _, found := m.Get(p); !found {
				return fmt.Errorf("missing required field '%v', expected %v", p, s)
			}
		}
	}

	if len(s.Enum) > 0 {
		return checkValueIsInEnum(i, s.Enum, &schema.Schema{Type: schema.Types{"object"}})
	}

	return nil
}
