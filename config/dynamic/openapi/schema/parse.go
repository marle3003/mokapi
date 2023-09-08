package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mokapi/media"
	"mokapi/sortedmap"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func ParseString(s string, schema *Ref) (interface{}, error) {
	return parse(s, schema)
}

func Parse(b []byte, contentType media.ContentType, schema *Ref) (i interface{}, err error) {
	switch {
	case contentType.Subtype == "json":
		d := &data{}
		err = json.Unmarshal(b, &d)
		if err != nil {
			return nil, fmt.Errorf("invalid json format: %v", err)
		}
		i = d.d
	case contentType.IsXml():
		i, err = parseXml(b, schema)
	default:
		i = string(b)
	}

	if err != nil {
		return
	}

	i, err = parse(i, schema)
	return
}

func ParseFrom(r io.Reader, contentType media.ContentType, schema *Ref) (i interface{}, err error) {
	switch contentType.Subtype {
	case "json":
		d := &data{}
		err = json.NewDecoder(r).Decode(&d)
		if err != nil {
			return nil, fmt.Errorf("invalid json format: %v", err)
		}
		i = d.d
	default:
		i, err = ioutil.ReadAll(r)
	}

	if err != nil {
		return
	}

	return parse(i, schema)
}

func parse(v interface{}, schema *Ref) (interface{}, error) {
	if schema == nil || schema.Value == nil {
		if m, ok := v.(*sortedmap.LinkedHashMap); ok {
			return toObject(m)
		}
		return v, nil
	}

	if len(schema.Value.AnyOf) > 0 {
		return parseAnyOf(v, schema.Value)
	} else if len(schema.Value.AllOf) > 0 {
		return parseAllOf(v, schema.Value)
	} else if len(schema.Value.OneOf) > 0 {
		return parseOneOf(v, schema.Value)
	} else if len(schema.Value.Type) == 0 {
		// A schema without a type matches any data type
		if _, ok := v.(*sortedmap.LinkedHashMap); ok {
			return parseObject(v, &Schema{Type: "object"})
		}
		return v, nil
	}

	switch schema.Value.Type {
	case "object":
		return parseObject(v, schema.Value)
	case "array":
		return parseArray(v, schema.Value)
	case "boolean":
		return readBoolean(v, schema.Value)
	case "integer":
		return parseInteger(v, schema.Value)
	case "number":
		return parseNumber(v, schema.Value)
	case "string":
		return parseString(v, schema.Value)
	}

	return nil, fmt.Errorf("unsupported type %q", schema.Value.Type)
}

func parseAnyOf(i interface{}, schema *Schema) (interface{}, error) {
	if m, ok := i.(*sortedmap.LinkedHashMap); ok {
		return parseAnyObject(m, schema)
	}
	return parseAnyValue(i, schema)
}

func parseAnyObject(m *sortedmap.LinkedHashMap, schema *Schema) (interface{}, error) {
	fields := make([]reflect.StructField, 0, m.Len())
	values := make([]reflect.Value, 0, m.Len())

	for _, ref := range schema.AnyOf {
		s := ref.Value

		required := make(map[string]struct{})
		for _, r := range s.Required {
			required[r] = struct{}{}
		}

		// free-form object
		if s.Properties == nil {
			return toObject(m)
		}

		for it := s.Properties.Value.Iter(); it.Next(); {
			name := it.Key().(string)
			pRef := it.Value().(*Ref)

			if v := m.Get(name); v == nil {
				if _, ok := required[name]; ok && len(required) > 0 {
					return nil, fmt.Errorf("missing required property %v, expected %v", name, schema)
				}
				continue
			}

			v, err := parse(m.Get(name), pRef)
			if err != nil {
				continue
			}
			values = append(values, reflect.ValueOf(v))
			fields = append(fields, newField(name, v))
		}
	}

	if m.Len() > len(fields) {
		return nil, fmt.Errorf("could not parse %v, too many properties for object, expected %v", toString(m), schema)
	}

	if err := validFields(fields); err != nil {
		return nil, err
	}

	t := reflect.StructOf(fields)
	v := reflect.New(t).Elem()
	for i, val := range values {
		v.Field(i).Set(val)
	}
	return v.Addr().Interface(), nil
}

func parseAnyValue(i interface{}, schema *Schema) (interface{}, error) {
	for _, s := range schema.AnyOf {
		i, err := parse(i, s)
		if err == nil {
			return i, nil
		}
	}
	return nil, fmt.Errorf("could not parse %v, expected %v", i, schema)
}

func parseAllOf(i interface{}, schema *Schema) (interface{}, error) {
	m, ok := i.(*sortedmap.LinkedHashMap)
	if !ok {
		return nil, fmt.Errorf("could not parse %v as object, expected %v", toString(i), schema)
	}
	fields := make([]reflect.StructField, 0, m.Len())
	values := make([]reflect.Value, 0, m.Len())

	for _, sRef := range schema.AllOf {
		s := sRef.Value

		required := make(map[string]struct{})
		for _, r := range s.Required {
			required[r] = struct{}{}
		}

		// free-form object
		if s.Properties == nil {
			return toObject(m)
		}

		for it := s.Properties.Value.Iter(); it.Next(); {
			name := it.Key().(string)
			pRef := it.Value().(*Ref)

			if v := m.Get(name); v == nil {
				if _, ok := required[name]; ok && len(required) > 0 {
					return nil, fmt.Errorf("could not parse %v, missing required property %v, expected %v", toString(i), name, schema)
				}
				continue
			}

			v, err := parse(m.Get(name), pRef)
			if err != nil {
				return nil, fmt.Errorf("could not parse %v, value does not match all schema, expected %v", toString(i), schema)
			}
			values = append(values, reflect.ValueOf(v))
			fields = append(fields, newField(name, v))
		}
	}

	if err := validFields(fields); err != nil {
		return nil, err
	}

	t := reflect.StructOf(fields)
	v := reflect.New(t).Elem()
	for i, val := range values {
		v.Field(i).Set(val)
	}
	return v.Addr().Interface(), nil
}

func parseOneOf(i interface{}, schema *Schema) (interface{}, error) {
	if m, ok := i.(*sortedmap.LinkedHashMap); ok {
		return parseOneOfObject(m, schema)
	}
	return parseOneOfValue(i, schema)
}

func parseOneOfObject(m *sortedmap.LinkedHashMap, schema *Schema) (interface{}, error) {
	var result interface{}
	for _, sRef := range schema.OneOf {
		s := sRef.Value
		fields := make([]reflect.StructField, 0, m.Len())
		values := make([]reflect.Value, 0, m.Len())

		required := make(map[string]struct{})
		for _, r := range s.Required {
			required[r] = struct{}{}
		}

		for it := s.Properties.Value.Iter(); it.Next(); {
			name := it.Key().(string)
			pRef := it.Value().(*Ref)

			if v := m.Get(name); v == nil {
				if _, ok := required[name]; ok && len(required) > 0 {
					return nil, fmt.Errorf("could not parse %v, missing required property %v, expected %v", toString(m), name, schema)
				}
				continue
			}

			v, err := parse(m.Get(name), pRef)
			if err != nil {
				continue
			}
			values = append(values, reflect.ValueOf(v))
			fields = append(fields, newField(name, v))
		}

		if m.Len() > len(fields) {
			continue
		}

		if result != nil {
			return nil, fmt.Errorf("could not parse %v, it is not valid for only one schema, expected %v", toString(m), schema)
		}

		if err := validFields(fields); err != nil {
			return nil, err
		}

		t := reflect.StructOf(fields)
		v := reflect.New(t).Elem()
		for i, val := range values {
			v.Field(i).Set(val)
		}
		result = v.Addr().Interface()
	}

	if result == nil {
		return nil, fmt.Errorf("could not parse %v, expected %v", toString(m), schema)
	}

	return result, nil
}

func parseOneOfValue(i interface{}, schema *Schema) (interface{}, error) {
	var result interface{}
	for _, s := range schema.OneOf {
		v, err := parse(i, s)
		if err != nil {
			continue
		}
		if result != nil {
			return nil, fmt.Errorf("could not parse %v because it is not valid for only one, expected %v", i, schema)
		}
		result = v
	}

	return result, nil
}

func parseObject(i interface{}, s *Schema) (interface{}, error) {
	m, ok := i.(*sortedmap.LinkedHashMap)
	if !ok {
		return nil, fmt.Errorf("could not parse %v as object", toString(i))
	}

	var err error

	if err = validateObject(m, s); err != nil {
		return nil, err
	}

	if s.IsDictionary() {
		result := make(map[string]interface{})
		for it := m.Iter(); it.Next(); {
			k := it.Key().(string)
			v := it.Value()
			v, err = parse(v, s.AdditionalProperties.Ref)
			if err != nil {
				return nil, err
			}
			result[k] = v
		}
		return result, nil
	}

	if !s.HasProperties() {
		return toObject(m)
	}

	if m.Len() > s.Properties.Value.Len() {
		return nil, fmt.Errorf("could not parse %v, too many properties", toString(m))
	}

	fields := make([]reflect.StructField, 0, m.Len())
	values := make([]reflect.Value, 0, m.Len())

	for it := s.Properties.Value.Iter(); it.Next(); {
		name := it.Key().(string)
		pRef := it.Value().(*Ref)
		if v := m.Get(name); v == nil {
			continue
		}

		v, err := parse(m.Get(name), pRef)
		if err != nil {
			return nil, err
		}
		values = append(values, reflect.ValueOf(v))
		fields = append(fields, newField(name, v))
	}

	for it := m.Iter(); it.Next(); {
		k := it.Key().(string)
		v := it.Value()
		if p := s.Properties.Get(k); p != nil {
			continue
		}
		if child, ok := v.(*sortedmap.LinkedHashMap); ok {
			v, err = toObject(child)
			if err != nil {
				return nil, err
			}
		}
		fields = append(fields, newField(k, v))
		values = append(values, reflect.ValueOf(v))
	}

	if err := validFields(fields); err != nil {
		return nil, err
	}

	t := reflect.StructOf(fields)
	v := reflect.New(t).Elem()
	for i, val := range values {
		v.Field(i).Set(val)
	}

	o := v.Addr().Interface()
	return o, validateObject(o, s)
}

func parseArray(i interface{}, s *Schema) (interface{}, error) {
	var sliceOf reflect.Type
	switch s.Items.Value.Type {
	case "object":
		sliceOf = reflect.TypeOf([]interface{}{})
	default:
		t, err := getType(s.Items.Value)
		if err != nil {
			return nil, err
		}
		sliceOf = reflect.SliceOf(t)
	}
	result := reflect.MakeSlice(sliceOf, 0, 0)

	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Slice:
		for index := 0; index < v.Len(); index++ {
			item, err := parse(v.Index(index).Interface(), s.Items)
			if err != nil {
				return nil, err
			}
			result = reflect.Append(result, reflect.ValueOf(item))
		}
	}

	ret := result.Interface()
	return ret, validateArray(ret, s)
}

func parseInteger(i interface{}, s *Schema) (n int64, err error) {
	switch v := i.(type) {
	case int:
		n = int64(v)
	case int64:
		n = v
	case float64:
		if math.Trunc(v) != v {
			return 0, fmt.Errorf("could not parse %v as integer, expected %v", i, s)
		}
		n = int64(v)
	case int32:
		n = int64(v)
	case string:
		switch s.Format {
		case "int64":
			n, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("could not parse '%v' as int64, expected %v", i, s)
			}
			return n, nil
		default:
			temp, err := strconv.Atoi(v)
			if err != nil {
				return 0, fmt.Errorf("could not parse '%v' as int, expected %v", i, s)
			}
			n = int64(temp)
		}
	default:
		return 0, fmt.Errorf("could not parse '%v' as int, expected %v", i, s)
	}

	switch s.Format {
	case "int32":
		if n > math.MaxInt32 || n < math.MinInt32 {
			return 0, fmt.Errorf("could not parse '%v', represents a number either less than int32 min value or greater max value, expected %v", i, s)
		}
	}

	return n, validateInt64(n, s)
}

func parseNumber(i interface{}, s *Schema) (f float64, err error) {
	switch v := i.(type) {
	case float64:
		f = v
	case string:
		f, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("could not parse '%v' as floating number, expected %v", i, s)
		}
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	default:
		return 0, fmt.Errorf("could not parse '%v' as floating number, expected %v", v, s)
	}

	switch s.Format {
	case "float":
		if f > math.MaxFloat32 {
			return 0, fmt.Errorf("could not parse %v as float, expected %v", i, s)
		}
	}

	return f, validateFloat64(f, s)
}

func parseString(v interface{}, schema *Schema) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("could not parse %v as string, expected %v", v, schema)
	}

	return s, validateString(s, schema)
}

func readBoolean(i interface{}, s *Schema) (bool, error) {
	if b, ok := i.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("could not parse %v as boolean, expected %v", i, s)
}

func toObject(m *sortedmap.LinkedHashMap) (interface{}, error) {
	fields := make([]reflect.StructField, 0, m.Len())
	values := make([]reflect.Value, 0, m.Len())

	for it := m.Iter(); it.Next(); {
		k := it.Key().(string)
		v := it.Value()
		if child, ok := v.(*sortedmap.LinkedHashMap); ok {
			var err error
			v, err = toObject(child)
			if err != nil {
				return nil, err
			}
		}

		fields = append(fields, newField(k, v))
		values = append(values, reflect.ValueOf(v))
	}

	if err := validFields(fields); err != nil {
		return nil, err
	}

	t := reflect.StructOf(fields)
	v := reflect.New(t).Elem()
	for i, val := range values {
		v.Field(i).Set(val)
	}
	return v.Addr().Interface(), nil
}

func toString(i interface{}) string {
	//b, err := json.Marshal(i)
	//if err != nil {
	//	log.Errorf("error in schema.toString(): %v", err)
	//}
	//return string(b)
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
	case *sortedmap.LinkedHashMap:
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

func newField(name string, value interface{}) reflect.StructField {
	return reflect.StructField{
		Name: strings.Title(getValidFieldName(name)),
		Type: reflect.TypeOf(value),
		Tag:  reflect.StructTag(fmt.Sprintf(`json:"%v"`, name)),
	}
}

func validFields(fields []reflect.StructField) error {
	m := make(map[string]bool)
	for _, f := range fields {
		if _, ok := m[f.Name]; !ok {
			m[f.Name] = true
		} else {
			return fmt.Errorf("duplicate field name %v", f.Name)
		}
	}
	return nil
}

func getValidFieldName(fieldName string) string {
	var sb strings.Builder
	for i, c := range fieldName {
		if i == 0 && !isLetter(c) {
			continue
		}

		if !(isLetter(c) || unicode.IsDigit(c)) {
			sb.WriteRune('_')
		} else {
			sb.WriteRune(c)
		}
	}

	return sb.String()
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

// to ensure order in json unmarshal
// using sorted map instead of builtin map
type data struct {
	d interface{}
}

func (d *data) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		d.d = token
		return nil
	}
	if delim != '{' && delim != '[' {
		return fmt.Errorf("unexpected token %s; expected '{'", token)
	}

	if delim == '[' {
		a := make([]interface{}, 0)
		for dec.More() {
			val := &data{}
			err = dec.Decode(&val)
			if err != nil {
				return err
			}

			a = append(a, val.d)
		}
		d.d = a
		return nil
	}

	m := sortedmap.NewLinkedHashMap()
	d.d = m
	for {
		token, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}
		key := token.(string)
		val := &data{}
		err = dec.Decode(&val)
		if err != nil {
			return err
		}

		m.Set(key, val.d)
	}

}
