package encoding

import (
	"fmt"
	"math"
	"mokapi/config/dynamic/openapi"
	"reflect"
	"strings"
)

func readObject(i interface{}, s *openapi.Schema) (interface{}, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object but got %T", i)
	}

	required := make(map[string]struct{})
	for _, r := range s.Required {
		required[r] = struct{}{}
	}

	// free-form object
	if s.Properties == nil {
		return toObject(m), nil
	}

	if len(m) > s.Properties.Value.Len() {
		return nil, fmt.Errorf("too many properties for object")
	}

	fields := make([]reflect.StructField, 0, len(m))
	values := make([]reflect.Value, 0, len(m))

	for it := s.Properties.Value.Iter(); it.Next(); {
		name := it.Key().(string)
		pRef := it.Value().(*openapi.SchemaRef)
		if _, ok := m[name]; !ok {
			if _, ok := required[name]; ok && len(required) > 0 {
				return nil, fmt.Errorf("expected required property %v", name)
			}
			continue
		}

		v, err := parse(m[name], pRef)
		if err != nil {
			return nil, err
		}
		values = append(values, reflect.ValueOf(v))
		fields = append(fields, reflect.StructField{
			Name: strings.Title(name),
			Type: reflect.TypeOf(v),
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%v"`, name)),
		})
	}

	t := reflect.StructOf(fields)
	v := reflect.New(t).Elem()
	for i, val := range values {
		v.Field(i).Set(val)
	}
	return v.Addr().Interface(), nil
}

func readArray(i interface{}, s *openapi.Schema) (interface{}, error) {
	a, ok := i.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected array but got %T", i)
	}

	if len(a) == 0 {
		var sliceOf reflect.Type
		switch s.Items.Value.Type {
		case "object":
			sliceOf = reflect.TypeOf([]interface{}{})
		default:
			sliceOf = reflect.SliceOf(getType(s.Items.Value))
		}
		return reflect.MakeSlice(sliceOf, 0, 0).Interface(), nil
	}

	v, err := parse(a[0], s.Items)
	if err != nil {
		return nil, err
	}
	result := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(v)), 0, len(a))

	for _, v := range a {
		v, err := parse(v, s.Items)
		if err != nil {
			return nil, err
		}
		result = reflect.Append(result, reflect.ValueOf(v))
	}
	return result.Interface(), nil
}

func readInteger(i interface{}, s *openapi.Schema) (int64, error) {
	f, ok := i.(float64)
	if !ok {
		return 0, fmt.Errorf("expected integer got %T", i)
	}
	n := int64(f)
	if math.Trunc(f) != f {
		return 0, fmt.Errorf("expected integer but got floating number")
	}

	switch s.Format {
	case "int32":
		if n > math.MaxInt32 || n < math.MinInt32 {
			return 0, fmt.Errorf("integer is not int32")
		}
	}

	return n, validateInt64(n, s)
}
func readNumber(v interface{}, s *openapi.Schema) (float64, error) {
	f, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("expected float got %T", v)
	}

	switch s.Format {
	case "float32":
		if f > math.MaxFloat32 || f < -math.MaxFloat32 {
			return 0, fmt.Errorf("number is not float32")
		}
	}

	return f, validateFloat64(f, s)
}
func readString(v interface{}, schema *openapi.SchemaRef) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("expected string got %T", v)
	}
	return s, nil
}
func readBoolean(i interface{}, _ *openapi.SchemaRef) (bool, error) {
	if b, ok := i.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("expected bool but got %T", i)
}
