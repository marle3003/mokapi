package mustache

import (
	"fmt"
	"mokapi/sortedmap"
	"reflect"
	"strings"
)

type Resolver interface {
	Resolve(name interface{}) (interface{}, error)
}

func Resolve(selector interface{}, i interface{}) (interface{}, error) {
	switch s := selector.(type) {
	case string:
		segments := strings.Split(s, ".")
		var err error
		for _, seg := range segments {
			i, err = resolveMember(seg, i)
			if err != nil {
				return i, err
			}
		}
	default:
		return resolveMember(selector, i)
	}

	return i, nil
}

func resolveMember(name interface{}, i interface{}) (interface{}, error) {
	if i == nil {
		return nil, fmt.Errorf("null reference: can not resolve %q", name)
	}

	if r, ok := i.(Resolver); ok {
		return r.Resolve(name)
	} else if m, ok := i.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
		v, _ := m.Get(name.(string))
		return v, nil
	}

	v := reflect.ValueOf(i)
	var ptr reflect.Value
	if v.Type().Kind() == reflect.Ptr {
		ptr = v
		v = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(i))
		temp := ptr.Elem()
		temp.Set(v)
	}

	if v.Kind() == reflect.Map {
		if s, ok := name.(string); ok {
			return resolveMapMember(s, v)
		} else {
			return nil, fmt.Errorf("expected string but was %T: %v", name, name)
		}
	} else if v.Kind() == reflect.Struct {
		if s, ok := name.(string); ok {
			return resolveStructMember(s, v)
		} else {
			return nil, fmt.Errorf("expected string but was %T: %v", name, name)
		}
	} else if v.Kind() == reflect.Slice {
		if n, ok := name.(int); ok {
			return v.Index(n).Interface(), nil
		} else {
			return nil, fmt.Errorf("expected int but was %T: %v", name, name)
		}
	}

	return nil, fmt.Errorf("undefined field %q", name)
}

func resolveStructMember(name string, v reflect.Value) (interface{}, error) {
	if name == "*" {
		r := make([]interface{}, 0, v.Len())

		for i := 0; i < v.NumField(); i++ {
			r = append(r, v.Field(i).Interface())
		}

		return r, nil
	}

	fieldName := strings.Title(name)

	f := v.FieldByName(fieldName)
	if f.IsValid() {
		return f.Interface(), nil
	}

	return nil, fmt.Errorf("undefined field %q", name)
}

func resolveMapMember(name string, v reflect.Value) (interface{}, error) {
	if name == "*" {
		r := make([]interface{}, 0, v.Len())

		for _, k := range v.MapKeys() {
			r = append(r, v.MapIndex(k).Interface())
		}
		return r, nil
	}

	for _, k := range v.MapKeys() {
		if k.Kind() != reflect.String {
			return nil, fmt.Errorf("unsupported map key type %q", k.Kind())
		}
		if k.String() == name {
			return v.MapIndex(k).Interface(), nil
		}
	}

	return nil, fmt.Errorf("undefined field %q", name)
}
