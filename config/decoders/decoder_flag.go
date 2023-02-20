package decoders

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/sortedmap"
	"reflect"
	"strconv"
	"strings"
)

type FlagDecoder struct {
}

func (f *FlagDecoder) Decode(flags map[string][]string, element interface{}) error {
	for name, value := range flags {
		paths := strings.Split(name, ".")
		err := setValue(paths, value, reflect.ValueOf(element))
		if err != nil {
			return errors.Wrapf(err, "configuration error %v", name)
		}
	}

	return nil
}

func setValue(paths []string, value []string, element reflect.Value) error {
	if len(value) == 0 {
		return nil
	}

	//currentElement := reflect.ValueOf(element)

	/*for _, fieldName := range path {
		k := currentElement.Kind()
		if k != reflect.Struct {
			currentElement = currentElement.Elem().FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == strings.ToLower(fieldName) })
		} else {
			currentElement = currentElement.FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == strings.ToLower(fieldName) })
		}
		if !currentElement.IsValid() {
			return fmt.Errorf("no configuration entry found: %v", name)
		}
	}*/

	switch element.Kind() {
	case reflect.Struct:
		if element.Type() == reflect.TypeOf(sortedmap.LinkedHashMap{}) {
			ptr := reflect.New(element.Type())
			if err := setSortedMap(paths, value, ptr.Interface().(*sortedmap.LinkedHashMap)); err != nil {
				return err
			}
			element.Set(ptr.Elem())
			return nil
		} else {
			element = element.FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == strings.ToLower(paths[0]) })
			if !element.IsValid() {
				return fmt.Errorf("configuration not found")
			}
			return setValue(paths[1:], value, element)
		}
	case reflect.Pointer:
		if element.IsNil() {
			element.Set(reflect.New(element.Type().Elem()))
		}
		return setValue(paths, value, element.Elem())
	case reflect.String:
		if len(value) > 1 {
			return fmt.Errorf("unexpected list for key %v", paths[0])
		}
		element.SetString(value[0])
		return nil
	case reflect.Bool:
		if len(value) > 1 {
			return fmt.Errorf("unexpected list for key %v", paths[0])
		}
		b, err := strconv.ParseBool(value[0])
		if err != nil {
			return fmt.Errorf("value %v cannot be parsed as bool: %v", value, err.Error())
		}
		element.SetBool(b)
		return nil
	case reflect.Slice:
		if len(value) > 1 {
			element.Set(reflect.ValueOf(value))
		} else if len(value) == 1 {
			values := strings.FieldsFunc(value[0], func(r rune) bool {
				return r == ' ' || r == ','
			})
			for _, v := range values {
				element.Set(reflect.Append(element, reflect.ValueOf(v)))
			}
		}
		return nil
	case reflect.Map:
		if element.IsNil() {
			element.Set(reflect.MakeMap(element.Type()))
		}
		for _, v := range value {
			kv := strings.Split(v, "=")
			if len(kv) != 2 {
				return fmt.Errorf("value %v can not be parse for %v", value, paths[0])
			}

			key := reflect.ValueOf(kv[0])
			ptr := reflect.New(element.Type().Elem())
			if !element.MapIndex(key).IsValid() {
				element.SetMapIndex(key, ptr.Elem())
			} else {
				ptr.Elem().Set(element.MapIndex(key))
			}
			if err := setValue([]string{}, []string{kv[1]}, ptr); err != nil {
				return err
			}
			element.SetMapIndex(key, ptr.Elem())
		}
		return nil
	}

	panic(fmt.Errorf("unsupported config type: %v", element.Kind()))
}

func setSortedMap(paths []string, value []string, m *sortedmap.LinkedHashMap) error {
	for _, v := range value {
		kv := strings.Split(v, "=")
		if len(kv) != 2 {
			return fmt.Errorf("value %v can not be parse for %v", value, paths[0])
		}
		m.Set(kv[0], kv[1])
	}
	return nil
}
