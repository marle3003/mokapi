package parser

import (
	"errors"
	"fmt"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"reflect"
	"regexp"
	"regexp/syntax"
	"sort"
	"strings"
)

type AdditionalPropertiesNotAllowed struct {
	Properties []string
	Schema     *schema.Schema
}

func (e *AdditionalPropertiesNotAllowed) Error() string {
	var sb strings.Builder
	for _, prop := range e.Properties {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("property '%s' not defined and the schema does not allow additional properties", prop))
	}

	return sb.String()
}

func (p *Parser) parseObject(data interface{}, s *schema.Schema, evaluated map[string]bool) (*sortedmap.LinkedHashMap[string, interface{}], error) {
	var result *sortedmap.LinkedHashMap[string, interface{}]
	var err error

	if m, ok := data.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
		result, err = p.parseLinkedMap(m, s, evaluated)
	} else {

		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Struct:
			result, err = p.parseStruct(v, s, evaluated)
		case reflect.Map:
			result, err = p.parseMap(v, s, evaluated)
		default:
			return nil, Errorf("type", "invalid type, expected object but got %v", toType(data))
		}
	}

	return result, err
}

func (p *Parser) parseLinkedMap(m *sortedmap.LinkedHashMap[string, interface{}], s *schema.Schema, evaluated map[string]bool) (*sortedmap.LinkedHashMap[string, interface{}], error) {
	obj := sortedmap.NewLinkedHashMap()
	for it := m.Iter(); it.Next(); {
		name := it.Key()

		var field *schema.Ref
		if s.Properties != nil {
			field = s.Properties.Get(name)
		}

		if field != nil || s.IsFreeForm() {
			d, err := p.parse(it.Value(), field)
			if err != nil {
				return nil, err
			}
			obj.Set(name, d)
		}
	}

	err := p.validateObject(obj, s)

	return obj, err
}

func (p *Parser) parseStruct(v reflect.Value, s *schema.Schema, evaluated map[string]bool) (*sortedmap.LinkedHashMap[string, interface{}], error) {
	t := v.Type()
	obj := sortedmap.NewLinkedHashMap()
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		name := unTitle(ft.Name)
		tag := ft.Tag.Get("json")
		if len(tag) > 0 {
			name = strings.Split(tag, ",")[0]
		}
		val := v.Field(i)

		if prop := s.Properties.Get(name); prop != nil || s.IsFreeForm() {
			d, err := p.parse(val.Interface(), prop)
			if err != nil {
				return nil, err
			}
			obj.Set(name, d)
		}
	}

	err := p.validateObject(obj, s)

	return obj, err
}

func (p *Parser) parseMap(v reflect.Value, s *schema.Schema, evaluated map[string]bool) (*sortedmap.LinkedHashMap[string, interface{}], error) {
	obj := sortedmap.NewLinkedHashMap()
	var err PathErrors

	if s.HasProperties() {
		for it := s.Properties.Iter(); it.Next(); {
			name := it.Key()
			o := v.MapIndex(reflect.ValueOf(name))
			if !o.IsValid() {
				continue
			}
			d, valErr := p.parse(o.Interface(), it.Value())
			if valErr != nil {
				err = append(err, wrapError(name, valErr))
			}
			obj.Set(name, d)
			evaluated[name] = true
		}
	}

	if len(s.PatternProperties) > 0 {
		for _, vk := range v.MapKeys() {
			key := vk.String()
			value := v.MapIndex(vk)
			if pErr := p.ValidatePatternProperty(key, value.Interface(), s.PatternProperties); pErr != nil {
				err = append(err, wrapError("patternProperties", pErr))
			}
			evaluated[key] = true
		}
	}

	if !s.IsFreeForm() {
		var additionalProps []string
		for _, k := range v.MapKeys() {
			name := fmt.Sprintf("%v", k.Interface())
			if _, found := obj.Get(name); !found {
				if s.AdditionalProperties.IsFalse() {
					if !p.ValidateAdditionalProperties {
						continue
					}
					additionalProps = append(additionalProps, name)
					continue
				}

				o := v.MapIndex(k)
				d, valErr := p.parse(o.Interface(), s.AdditionalProperties)
				if valErr != nil {
					err = append(err, wrapError("additionalProperties", wrapError(name, valErr)))
				}
				obj.Set(name, d)
				evaluated[name] = true
			}
		}
		if len(additionalProps) > 0 {
			sort.Strings(additionalProps)
			err = append(err, wrapError("additionalProperties", &AdditionalPropertiesNotAllowed{Properties: additionalProps}))
		}
	} else {
		for _, k := range v.MapKeys() {
			name := fmt.Sprintf("%v", k.Interface())
			if _, found := obj.Get(name); !found {
				o := v.MapIndex(k)
				obj.Set(name, o.Interface())
			}
		}
	}

	valErr := p.validateObject(obj, s)
	if valErr != nil {
		err = append(err, valErr)
	}

	if len(err) > 0 {
		return obj, &err
	}

	return obj, nil
}

func (p *Parser) validateObject(obj *sortedmap.LinkedHashMap[string, interface{}], s *schema.Schema) error {
	var err PathErrors

	if s.MinProperties != nil && obj.Len() < *s.MinProperties {
		err = append(err, Errorf("minProperties", "property count %v is less than minimum count of %v", obj.Len(), *s.MinProperties))
	}
	if s.MaxProperties != nil && obj.Len() > *s.MaxProperties {
		err = append(err, Errorf("maxProperties", "property count %v exceeds maximum count of %v", obj.Len(), *s.MaxProperties))
	}

	if s.Required != nil {
		required := []string{}
		for _, p := range s.Required {
			if _, found := obj.Get(p); !found {
				required = append(required, p)
			}
		}
		if len(required) > 0 {
			err = append(err, Errorf("required", "required properties are missing: %v", strings.Join(required, ", ")))
		}
	}

	if s.PropertyNames != nil {
		for it := obj.Iter(); it.Next(); {
			name := it.Key()
			_, propErr := p.parse(name, s.PropertyNames)
			if propErr != nil {
				err = append(err, wrapError("propertyNames", propErr))
			}
		}
	}

	if len(s.Enum) > 0 {
		if enumErr := checkValueIsInEnum(obj, s.Enum, &schema.Schema{Type: schema.Types{"object"}}); enumErr != nil {
			err = append(err, enumErr)
		}
	}

	if len(err) > 0 {
		return &err
	}

	return nil
}

func (p *Parser) ValidatePatternProperty(name string, value interface{}, patternProperties map[string]*schema.Ref) error {
	for pattern, s := range patternProperties {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			var sErr *syntax.Error
			var msg string
			if errors.As(err, &sErr) {
				msg = sErr.Code.String()
			} else {
				msg = err.Error()
			}
			return wrapError(pattern, fmt.Errorf("validate string '%s' with regex pattern '%s' failed: error parsing regex: %s", name, pattern, msg))
		}
		if regex.MatchString(name) {
			if _, err := p.parse(value, s); err != nil {
				return wrapError(pattern, err)
			}
		}
	}
	return nil
}
