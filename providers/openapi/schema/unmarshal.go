package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	jsonSchema "mokapi/json/schema"
	"mokapi/media"
	"mokapi/sortedmap"
	"reflect"
	"strconv"
	"strings"
)

func ParseString(s string, schema *Ref) (interface{}, error) {
	p := parser{convertStringToNumber: true, convertStrintToBoolean: true}
	return p.parse(s, schema)
}

func UnmarshalFrom(r io.Reader, contentType media.ContentType, schema *Ref) (i interface{}, err error) {
	p := parser{}
	switch contentType.Subtype {
	case "json":
		err = json.NewDecoder(r).Decode(&i)
		if err != nil {
			return nil, fmt.Errorf("invalid json format: %v", err)
		}
	default:
		p.convertStringToNumber = true
		i, err = io.ReadAll(r)
	}

	if err != nil {
		return
	}

	return p.parse(i, schema)
}

type parser struct {
	convertStringToNumber  bool
	convertStrintToBoolean bool
	xml                    bool
}

func (p *parser) parse(v interface{}, schema *Ref) (interface{}, error) {
	if schema == nil || schema.Value == nil {
		return v, nil
	}

	if len(schema.Value.AnyOf) > 0 {
		return p.parseAny(v, schema.Value)
	}
	if len(schema.Value.AllOf) > 0 {
		return p.parseAllOf(v, schema.Value)
	}
	if len(schema.Value.OneOf) > 0 {
		return p.parseOneOf(v, schema.Value)
	}
	if len(schema.Value.Type) == 0 {
		// A schema without a type matches any data type
		if _, ok := v.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
			return p.parseObject(v, &Schema{Type: jsonSchema.Types{"object"}})
		}
		return v, nil
	}

	var err error
	for _, typeName := range schema.Value.Type {
		var result interface{}
		switch typeName {
		case "object":
			result, err = p.parseObject(v, schema.Value)
		case "array":
			result, err = p.parseArray(v, schema.Value)
		case "boolean":
			result, err = p.readBoolean(v, schema.Value)
		case "integer":
			result, err = p.parseInteger(v, schema.Value)
		case "number":
			result, err = p.parseNumber(v, schema.Value)
		case "string":
			result, err = p.parseString(v, schema.Value)
		default:
			err = fmt.Errorf("unsupported type %q", typeName)
		}
		if err == nil {
			return result, nil
		}
	}

	return nil, err
}

func (p *parser) parseAny(i interface{}, schema *Schema) (interface{}, error) {
	var result interface{}

	for _, ref := range schema.AnyOf {
		// free-form object
		if ref.Value.Type.Includes("object") && ref.Value.Properties == nil {
			return i, nil
		}

		part, err := p.parse(i, ref)
		if err != nil {
			continue
		}

		if values, ok := part.(map[string]interface{}); ok {
			m, ok := result.(map[string]interface{})
			if !ok {
				if result != nil {
					return result, nil
				} else {
					m = map[string]interface{}{}
				}
			}

			for k, v := range values {
				// only overwrite value if prop is in schema
				prop := ref.Value.Properties.Get(k)
				if _, found := m[k]; found && prop != nil {
					m[k] = v
				} else if !found {
					m[k] = v
				}
			}
			result = m
		} else {
			result = part
		}
	}

	if result == nil {
		return nil, fmt.Errorf("parse %v failed, expected %v", i, schema)
	}

	return result, nil
}

func (p *parser) parseAllOf(i interface{}, schema *Schema) (interface{}, error) {
	m := i.(map[string]interface{})

	result := map[string]interface{}{}
	for _, sRef := range schema.AllOf {
		s := sRef.Value

		// free-form object
		if s.Properties == nil {
			return m, nil
		}

		part, err := p.parse(m, sRef)
		if err != nil {
			return nil, fmt.Errorf("parse %v failed: value does not match part of allOf: %w", toString(i), err)
		}

		values, ok := part.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("ERROR")
		}
		for k, v := range values {
			// only overwrite value if prop is in schema
			prop := s.Properties.Get(k)
			if _, found := result[k]; found && prop != nil {
				result[k] = v
			} else if !found {
				result[k] = v
			}
		}
	}

	return result, nil
}

func (p *parser) parseOneOf(i interface{}, schema *Schema) (interface{}, error) {
	return p.parseOneOfObject(i, schema)
}

func (p *parser) parseOneOfObject(i interface{}, schema *Schema) (interface{}, error) {
	var result interface{}

	for _, ref := range schema.OneOf {
		// free-form object
		if ref.Value.Type.Includes("object") && ref.Value.Properties == nil {
			result = i
		}

		part, err := p.parse(i, ref)
		if err != nil {
			continue
		}

		if result != nil {
			return nil, fmt.Errorf("parse %v failed: it is valid for more than one schema, expected %v", toString(i), schema)
		}

		result = part
	}

	if result == nil {
		return nil, fmt.Errorf("parse %v failed: expected to match one of schema but it matches none", toString(i))
	}

	return result, nil
}

func (p *parser) parseObject(i interface{}, s *Schema) (interface{}, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("parse %v as object", toString(i))
	}

	var err error

	if err = validateObject(m, s); err != nil {
		return nil, err
	}

	if s.IsDictionary() {
		result := make(map[string]interface{})
		for k, v := range m {
			parsed, valError := p.parse(v, s.AdditionalProperties.Ref)
			if valError != nil {
				err = errors.Join(err, valError)
			}
			result[k] = parsed
		}
		return result, nil
	}

	if !s.HasProperties() {
		return m, nil
	}

	result := map[string]interface{}{}
	for k, v := range m {
		name, prop := p.getProperty(k, s.Properties)

		parsed, propError := p.parse(v, prop)
		if propError != nil {
			err = errors.Join(err, fmt.Errorf("parse '%v' failed: %w", k, propError))
		}
		result[name] = parsed
	}

	return result, err
}

func (p *parser) getProperty(name string, props *Schemas) (string, *Ref) {
	if !p.xml {
		return name, props.Get(name)
	}

	for it := props.Iter(); it.Next(); {
		k := it.Key()
		r := it.Value()
		if r.Value == nil && k == name {
			return name, r
		}
		if r.Value == nil {
			continue
		}
		if r.Value.Xml != nil && name == r.Value.Xml.Name {
			return it.Key(), r
		}
		if k == name {
			return name, r
		}
	}

	return "", nil
}

func (p *parser) parseArray(i interface{}, s *Schema) (interface{}, error) {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected array but got %v", toString(i))
	}

	var err error
	result := make([]interface{}, 0)
	for i := 0; i < v.Len(); i++ {
		o := v.Index(i)
		v, errItem := p.parse(o.Interface(), s.Items)
		if errItem != nil {
			err = errors.Join(err, errItem)
		}
		result = append(result, v)
	}

	if errVal := validateArray(result, s); errVal != nil {
		err = errors.Join(errVal)
	}

	return result, err
}

func (p *parser) parseInteger(i interface{}, s *Schema) (n int64, err error) {
	switch v := i.(type) {
	case int:
		n = int64(v)
	case int64:
		n = v
	case float64:
		if math.Trunc(v) != v {
			return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
		}
		n = int64(v)
	case int32:
		n = int64(v)
	case string:
		if !p.convertStringToNumber {
			return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
		}
		switch s.Format {
		case "int64":
			n, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
			}
			return n, nil
		default:
			temp, err := strconv.Atoi(v)
			if err != nil {
				return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
			}
			n = int64(temp)
		}
	default:
		return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
	}

	switch s.Format {
	case "int32":
		if n > math.MaxInt32 || n < math.MinInt32 {
			return 0, fmt.Errorf("parse '%v' failed: represents a number either less than int32 min value or greater max value, expected %v", i, s)
		}
	}

	return n, validateInt64(n, s)
}

func (p *parser) parseNumber(i interface{}, s *Schema) (f float64, err error) {
	switch v := i.(type) {
	case float64:
		f = v
	case string:
		if !p.convertStringToNumber {
			return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
		}
		f, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("parse '%v' failed, expected %v", i, s)
		}
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	default:
		return 0, fmt.Errorf("parse '%v' failed, expected %v", v, s)
	}

	switch s.Format {
	case "float":
		if f > math.MaxFloat32 {
			return 0, fmt.Errorf("parse %v failed, expected %v", i, s)
		}
	}

	return f, validateFloat64(f, s)
}

func (p *parser) parseString(v interface{}, schema *Schema) (interface{}, error) {
	s, ok := v.(string)
	if !ok {
		if schema.IsNullable() {
			return nil, nil
		}
		return nil, fmt.Errorf("parse %v failed, expected %v", v, schema)
	}

	return s, validateString(s, schema)
}

func (p *parser) readBoolean(i interface{}, s *Schema) (bool, error) {
	switch v := i.(type) {
	case bool:
		return v, nil
	case string:
		if p.convertStrintToBoolean {
			switch strings.ToLower(v) {
			case "true":
				return true, nil
			case "false":
				return false, nil
			}
		}
		return false, fmt.Errorf("parse %v failed, expected %v", i, s)
	case int, int64:
		return v != 0, nil
	}
	return false, fmt.Errorf("parse %v failed, expected %v", i, s)
}

func toString(i interface{}) string {
	var sb strings.Builder
	switch o := i.(type) {
	case []interface{}:
		sb.WriteRune('[')
		for i, v := range o {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(toString(v))
		}
		sb.WriteRune(']')
	case map[string]interface{}:
		sb.WriteRune('{')
		for key, val := range o {
			if sb.Len() > 1 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%v: %v", key, toString(val)))
		}
		sb.WriteRune('}')
	case string, int, int32, int64, float32, float64:
		sb.WriteString(fmt.Sprintf("%v", o))
	case *sortedmap.LinkedHashMap[string, interface{}]:
		return o.String()
	default:
		v := reflect.ValueOf(i)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		t := reflect.TypeOf(i)
		switch v.Kind() {
		case reflect.Slice:
			sb.WriteRune('[')
			for i := 0; i < v.Len(); i++ {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(toString(v.Index(i).Interface()))
			}
			sb.WriteRune(']')
		case reflect.Struct:
			sb.WriteRune('{')
			for i := 0; i < v.NumField(); i++ {
				if i > 0 {
					sb.WriteString(", ")
				}
				name := t.Field(i).Name
				fv := v.Field(i).Interface()
				sb.WriteString(fmt.Sprintf("%v: %v", name, fv))
			}
			sb.WriteRune('}')
		}
	}
	return sb.String()
}
