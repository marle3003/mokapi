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
			return nil, &ErrorDetail{
				Message: fmt.Sprintf("invalid type, expected object but got %v", toType(data)),
				Field:   "type",
			}
		}
	}

	return result, err
}

func (p *Parser) parseLinkedMap(m *sortedmap.LinkedHashMap[string, interface{}], s *schema.Schema, evaluated map[string]bool) (*sortedmap.LinkedHashMap[string, interface{}], error) {
	obj := sortedmap.NewLinkedHashMap()
	var err ErrorList

	if s.HasProperties() {
		for it := s.Properties.Iter(); it.Next(); {
			name := it.Key()
			prop := it.Value()
			v, ok := m.Get(it.Key())
			if !ok {
				if prop != nil && prop.Default != nil {
					obj.Set(name, prop.Default)
					evaluated[name] = true
				}
				continue
			}

			d, valErr := p.parse(v, prop)
			if valErr != nil {
				err = append(err, wrapErrorDetail(valErr, &ErrorDetail{
					Field: name,
				}))
			}
			obj.Set(name, d)
			evaluated[name] = true
		}
	}

	if len(s.PatternProperties) > 0 {
		for it := m.Iter(); it.Next(); {
			key := it.Key()
			value := it.Value()
			if pErr := p.ValidatePatternProperty(key, value, s.PatternProperties); pErr != nil {
				err = append(err, wrapErrorDetail(pErr, &ErrorDetail{
					Field: "patternProperties",
				}))
			}
			evaluated[key] = true
		}
	}

	if !s.IsFreeForm() {
		var additionalProps []string
		for it := m.Iter(); it.Next(); {
			name := fmt.Sprintf("%v", it.Key())
			if _, found := obj.Get(name); !found {
				if s.AdditionalProperties.IsFalse() {
					if !p.ValidateAdditionalProperties {
						continue
					}
					additionalProps = append(additionalProps, name)
					continue
				}

				d, valErr := p.parse(it.Value(), s.AdditionalProperties)
				if valErr != nil {
					err = append(err, wrapErrorDetail(valErr, &ErrorDetail{
						Field: fmt.Sprintf("additionalProperties/%s", name),
					}))
				}
				obj.Set(name, d)
				evaluated[name] = true
			}
		}
		if len(additionalProps) > 0 {
			sort.Strings(additionalProps)
			for _, prop := range additionalProps {
				err = append(err,
					&ErrorDetail{
						Field:   "additionalProperties",
						Message: fmt.Sprintf("property '%s' not defined and the schema does not allow additional properties", prop),
					})
			}
		}
	} else {
		for it := m.Iter(); it.Next(); {
			name := fmt.Sprintf("%v", it.Key())
			if _, found := obj.Get(name); !found {
				obj.Set(name, it.Value())
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
			evaluated[name] = true
		}
	}

	valErr := p.validateObject(obj, s)
	if valErr != nil {
		return nil, valErr
	}

	return obj, nil
}

func (p *Parser) parseMap(v reflect.Value, s *schema.Schema, evaluated map[string]bool) (*sortedmap.LinkedHashMap[string, interface{}], error) {
	obj := sortedmap.NewLinkedHashMap()
	var err ErrorList

	if s.HasProperties() {
		for it := s.Properties.Iter(); it.Next(); {
			name := it.Key()
			prop := it.Value()
			o := v.MapIndex(reflect.ValueOf(name))
			if !o.IsValid() {
				if prop != nil && prop.Default != nil {
					obj.Set(name, prop.Default)
					evaluated[name] = true
				}
				continue
			}
			d, valErr := p.parse(o.Interface(), prop)
			if valErr != nil {
				err = append(err, wrapErrorDetail(valErr, &ErrorDetail{
					Field: name,
				}))
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
				err = append(err, wrapErrorDetail(pErr, &ErrorDetail{
					Field: "patternProperties",
				}))
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
					err = append(err, wrapErrorDetail(valErr, &ErrorDetail{
						Field: fmt.Sprintf("additionalProperties/%s", name),
					}))
				}
				obj.Set(name, d)
				evaluated[name] = true
			}
		}
		if len(additionalProps) > 0 {
			sort.Strings(additionalProps)
			for _, prop := range additionalProps {
				err = append(err,
					&ErrorDetail{
						Field:   "additionalProperties",
						Message: fmt.Sprintf("property '%s' not defined and the schema does not allow additional properties", prop),
					})
			}
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
		err = append(err, *valErr...)
	}

	if len(err) > 0 {
		return obj, &err
	}

	return obj, nil
}

func (p *Parser) validateObject(obj *sortedmap.LinkedHashMap[string, interface{}], s *schema.Schema) *ErrorList {
	var err ErrorList

	if s.MinProperties != nil && obj.Len() < *s.MinProperties {
		err = append(err, &ErrorDetail{
			Message: fmt.Sprintf("property count %v is less than minimum count of %v", obj.Len(), *s.MinProperties),
			Field:   "minProperties",
		})
	}
	if s.MaxProperties != nil && obj.Len() > *s.MaxProperties {
		err = append(err, &ErrorDetail{
			Message: fmt.Sprintf("property count %v exceeds maximum count of %v", obj.Len(), *s.MaxProperties),
			Field:   "maxProperties",
		})
	}

	if s.Required != nil {
		required := []string{}
		for _, p := range s.Required {
			if _, found := obj.Get(p); !found {
				required = append(required, p)
			}
		}
		if len(required) > 0 {
			err = append(err, &ErrorDetail{
				Message: fmt.Sprintf("required properties are missing: %v", strings.Join(required, ", ")),
				Field:   "required",
			})
		}
	}

	if s.PropertyNames != nil {
		for it := obj.Iter(); it.Next(); {
			name := it.Key()
			_, propErr := p.parse(name, s.PropertyNames)
			if propErr != nil {
				err = append(err, wrapErrorDetail(propErr, &ErrorDetail{
					Field: "propertyNames",
				}))
			}
		}
	}

	for k, required := range s.DependentRequired {
		if _, found := obj.Get(k); found {
			var missing []string
			for _, req := range required {
				if _, found = obj.Get(req); !found {
					missing = append(missing, req)
				}
			}
			if len(missing) > 0 {
				err = append(err, &ErrorDetail{
					Message: fmt.Sprintf("dependencies for property '%v' failed: missing required keys: %v.", k, strings.Join(missing, ", ")),
					Field:   "dependentRequired",
				})
			}
		}
	}

	for k, required := range s.DependentSchemas {
		if _, found := obj.Get(k); found {
			_, reqErr := p.parse(obj, required)
			if reqErr != nil {
				propErr := wrapErrorDetail(reqErr, &ErrorDetail{
					Field: fmt.Sprintf("dependentSchemas/%s", k),
				})
				err = append(err, propErr)
			}
		}
	}

	if s.If != nil {
		if _, ifErr := p.parse(obj, s.If); ifErr == nil {
			if s.Then != nil {
				if _, thenErr := p.parse(obj, s.Then); thenErr != nil {
					err = append(err, wrapErrorDetail(thenErr, &ErrorDetail{
						Message: "does not match schema",
						Field:   "then",
					}))
				}
			}
		} else if s.Else != nil {
			if _, elseErr := p.parse(obj, s.Else); elseErr != nil {
				err = append(err, wrapErrorDetail(elseErr, &ErrorDetail{
					Message: "does not match schema",
					Field:   "else",
				}))
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

func (p *Parser) ValidatePatternProperty(name string, value interface{}, patternProperties map[string]*schema.Schema) error {
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
			return &ErrorDetail{
				Message: fmt.Sprintf("validate string '%s' with regex pattern '%s' failed: error parsing regex: %s", name, pattern, msg),
				Field:   pattern,
			}
		}
		if regex.MatchString(name) {
			if _, err := p.parse(value, s); err != nil {
				return wrapErrorDetail(err, &ErrorDetail{
					Field: pattern,
				})
			}
		}
	}
	return nil
}
