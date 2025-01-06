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
	var err PathErrors

	if s.HasProperties() {
		for it := s.Properties.Iter(); it.Next(); {
			name := it.Key()
			prop := it.Value()
			v, ok := m.Get(it.Key())
			if !ok {
				if prop.Value != nil && prop.Value.Default != nil {
					obj.Set(name, prop.Value.Default)
					evaluated[name] = true
				}
				continue
			}

			d, valErr := p.parse(v, prop)
			if valErr != nil {
				err = append(err, wrapError(name, valErr))
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
				err = append(err, wrapError("patternProperties", pErr))
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

	err := p.validateObject(obj, s)

	return obj, err
}

func (p *Parser) parseMap(v reflect.Value, s *schema.Schema, evaluated map[string]bool) (*sortedmap.LinkedHashMap[string, interface{}], error) {
	obj := sortedmap.NewLinkedHashMap()
	var err PathErrors

	if s.HasProperties() {
		for it := s.Properties.Iter(); it.Next(); {
			name := it.Key()
			prop := it.Value()
			o := v.MapIndex(reflect.ValueOf(name))
			if !o.IsValid() {
				if prop.Value != nil && prop.Value.Default != nil {
					obj.Set(name, prop.Value.Default)
					evaluated[name] = true
				}
				continue
			}
			d, valErr := p.parse(o.Interface(), prop)
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

	for k, required := range s.DependentRequired {
		if _, found := obj.Get(k); found {
			var missing []string
			for _, req := range required {
				if _, found = obj.Get(req); !found {
					missing = append(missing, req)
				}
			}
			if len(missing) > 0 {
				err = append(err, Errorf("dependentRequired", "dependencies for property '%v' failed: missing required keys: %v.", k, strings.Join(missing, ", ")))
			}
		}
	}

	for k, required := range s.DependentSchemas {
		if _, found := obj.Get(k); found {
			_, reqErr := p.parse(obj, required)
			if reqErr != nil {
				pe := &PathCompositionError{
					Path:    k,
					Message: fmt.Sprintf("dependencies for property '%v' failed", k),
				}
				pe.append(reqErr)
				err = append(err, wrapError("dependentSchemas", pe))
			}
		}
	}

	if s.If != nil {
		if _, ifErr := p.parse(obj, s.If); ifErr == nil {
			if s.Then != nil {
				if _, thenErr := p.parse(obj, s.Then); thenErr != nil {
					pe := &PathCompositionError{
						Path:    "then",
						Message: fmt.Sprintf("does not match schema from 'then'"),
					}
					pe.append(thenErr)
					err = append(err, pe)
				}
			}
		} else if s.Else != nil {
			if _, elseErr := p.parse(obj, s.Else); elseErr != nil {
				pe := &PathCompositionError{
					Path:    "else",
					Message: fmt.Sprintf("does not match schema from 'else'"),
				}
				pe.append(elseErr)
				err = append(err, pe)
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
