package decoders

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type FlagDecoder struct {
}

func (f *FlagDecoder) Decode(flags map[string]string, element interface{}) error {
	keys := make([]string, 0, len(flags))
	for k := range flags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		value := flags[name]
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
	case reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("parse int64 failed: %v", err)
		}
		element.SetInt(i)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("value %v cannot be parsed as bool: %v", value, err.Error())
		}
		element.SetBool(b)
		return nil
	case reflect.Slice:
		return setArray(paths, value, element)
	case reflect.Map:
		return setMap(paths, value, element)
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

func setArray(paths []string, value string, element reflect.Value) error {
	if len(paths) > 0 {
		index, err := parseArrayIndex(paths[0])
		if err != nil {
			return fmt.Errorf("parse array index failed: %v", err)
		}

		if index >= element.Cap() {
			n := index + 1
			nCap := 2 * n
			if nCap < 4 {
				nCap = 4
			}
			if element.IsNil() {
				s := reflect.MakeSlice(element.Type(), n, nCap)
				element.Set(s)
			} else {
				s := reflect.MakeSlice(element.Type(), n, nCap)
				reflect.Copy(s, element)
				element.Set(s)
			}
		}
		if index >= element.Len() {
			element.SetLen(index + 1)
		}

		return setValue(paths[1:], value, element.Index(index))
	} else {
		values := strings.Split(value, ",")
		for _, v := range values {
			ptr := reflect.New(reflect.PointerTo(element.Type().Elem()))
			if err := setValue(paths, v, ptr); err != nil {
				return err
			}
			element.Set(reflect.Append(element, ptr.Elem().Elem()))
		}
	}

	return nil
}

func parseArrayIndex(path string) (int, error) {
	if strings.HasPrefix(path, "[") {
		s := strings.TrimPrefix(path, "[")
		s = strings.TrimSuffix(s, "]")
		return strconv.Atoi(s)

	}
	return strconv.Atoi(path)
}

func setMap(paths []string, value string, element reflect.Value) error {
	if element.IsNil() {
		element.Set(reflect.MakeMap(element.Type()))
	}

	key := reflect.ValueOf(paths[0])
	var ptr reflect.Value
	ptr = reflect.New(reflect.PointerTo(element.Type().Elem()))

	if element.MapIndex(key).IsValid() {
		ptr.Elem().Set(reflect.New(element.Type().Elem()))
		ptr.Elem().Elem().Set(element.MapIndex(key))
	}
	if err := setValue(paths[1:], value, ptr); err != nil {
		return err
	}

	element.SetMapIndex(key, ptr.Elem().Elem())

	return nil
}
