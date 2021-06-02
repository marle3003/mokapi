package objectpath

import (
	"fmt"
	"reflect"
	"strings"
)

type Resolver interface {
	Resolve(name string) (interface{}, error)
}

func Resolve(path string, i interface{}) (interface{}, error) {
	segments := strings.Split(path, ".")
	var err error
	for _, seg := range segments {
		i, err = resolveMember(seg, i)
		if err != nil {
			return i, err
		}
	}

	return i, nil
}

func resolveMember(name string, i interface{}) (interface{}, error) {
	if r, ok := i.(Resolver); ok {
		return r.Resolve(name)
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
		for _, k := range v.MapKeys() {
			if k.Kind() != reflect.String {
				return nil, fmt.Errorf("unsupported map key type %q", k.Kind())
			}
			if k.String() == name {
				return v.MapIndex(k).Interface(), nil
			}
		}
	} else if v.Kind() == reflect.Struct {

		fieldName := strings.Title(name)

		f := v.FieldByName(fieldName)
		if !f.IsValid() {
			// check for field on pointer
			f = reflect.Indirect(ptr).FieldByName(fieldName)
		}
		if f.IsValid() {
			return f.Interface(), nil
		}
	}

	return nil, fmt.Errorf("undefined field %q", name)
}
