package decoders

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

type FlagDecoder struct {
}

func (f *FlagDecoder) Decode(flags map[string]string, element interface{}) error {
	for name, value := range flags {
		paths := parsePath(name)
		err := setValue(paths, value, reflect.ValueOf(element))
		if err != nil {
			return errors.Wrapf(err, "configuration error %v", name)
		}
	}

	return nil
}

func setValue(paths []string, value string, element reflect.Value) error {
	if len(value) == 0 {
		return nil
	}

	switch element.Kind() {
	case reflect.Struct:
		element = element.FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == strings.ToLower(paths[0]) })
		if !element.IsValid() {
			return fmt.Errorf("configuration not found")
		}
		return setValue(paths[1:], value, element)
	case reflect.Pointer:
		if element.IsNil() {
			element.Set(reflect.New(element.Type().Elem()))
		}
		return setValue(paths, value, element.Elem())
	case reflect.String:
		element.SetString(value)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("value %v cannot be parsed as bool: %v", value, err.Error())
		}
		element.SetBool(b)
		return nil
	case reflect.Slice:
		ptr := reflect.New(reflect.PointerTo(element.Type().Elem()))
		if err := setValue(paths[1:], value, ptr); err != nil {
			return err
		}
		element.Set(reflect.Append(element, ptr.Elem().Elem()))
		return nil
	case reflect.Map:
		if element.IsNil() {
			element.Set(reflect.MakeMap(element.Type()))
		}

		key := reflect.ValueOf(paths[0])
		ptr := reflect.New(reflect.PointerTo(element.Type().Elem()))
		if err := setValue(paths[1:], value, ptr); err != nil {
			return err
		}

		element.SetMapIndex(key, ptr.Elem().Elem())

		return nil
	}

	panic(fmt.Errorf("unsupported config type: %v", element.Kind()))
}

func parsePath(key string) []string {
	var paths []string
	split := strings.Split(key, ".")
	for _, v := range split {
		if strings.HasSuffix(v, "]") {
			index := strings.Index(v, "[")
			paths = append(paths, v[:index], v[index:])
		} else {
			paths = append(paths, v)
		}
	}
	return paths
}
