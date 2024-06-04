package decoders

import (
	"encoding/json"
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
		field := element.FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == strings.ToLower(paths[0]) })
		if !field.IsValid() {
			err := explode(element, paths[0], value)
			if err != nil {
				return fmt.Errorf("configuration not found")
			}
			return nil
		}

		k := field.Type().Kind()
		if len(paths) == 1 && (k == reflect.Struct || k == reflect.Slice) {
			p := reflect.New(field.Type())
			p.Elem().Set(field)
			i, err := convert(value, p)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(i).Elem())
			return nil
		} else {
			return setValue(paths[1:], value, field)
		}
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

func explode(v reflect.Value, name string, value string) error {
	f := getFieldByTag(v, name)
	if !f.IsValid() {
		return fmt.Errorf("not found")
	}

	p := reflect.New(f.Type().Elem())
	i, err := convertJson(value, p)
	if err != nil {
		return err
	}
	f.Set(reflect.Append(f, reflect.ValueOf(i).Elem()))
	return nil
}

func getFieldByTag(v reflect.Value, name string) reflect.Value {
	for i := 0; i < v.NumField(); i++ {
		explode := v.Type().Field(i).Tag.Get("explode")
		if explode == name {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}

func convert(s string, v reflect.Value) (interface{}, error) {
	i, err := convertJson(s, v)
	if err == nil {
		return i, nil
	}

	if v.Elem().Type().Kind() == reflect.Struct {
		pairs := strings.Split(s, ",")
		for _, pair := range pairs {
			kv := strings.Split(pair, "=")
			if len(kv) != 2 {
				return nil, fmt.Errorf("parse shorthand failed: %v", s)
			}
			field := v.Elem().FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == strings.ToLower(kv[0]) })
			if !field.IsValid() {
				return nil, fmt.Errorf("field %v not found", kv[0])
			}
			if field.Type().Kind() != reflect.String {
				return s, fmt.Errorf("not supported shorthand with nested objects")
			}

			field.Set(reflect.ValueOf(kv[1]))
		}
	} else if v.Elem().Type().Kind() == reflect.Slice {
		items := strings.Split(s, " ")
		for _, item := range items {
			p := reflect.New(v.Elem().Type().Elem())
			i, err := convert(item, p)
			if err != nil {
				return nil, err
			}
			vItem := reflect.ValueOf(i)
			if vItem.Type().Kind() == reflect.Pointer {
				vItem = vItem.Elem()
			}
			v.Elem().Set(reflect.Append(v.Elem(), vItem))
		}
	} else if v.Elem().Type().Kind() == reflect.String {
		return s, nil
	}

	return v.Interface(), nil
}

func convertJson(s string, v reflect.Value) (interface{}, error) {
	i := v.Interface()
	err := json.Unmarshal([]byte(s), i)
	return i, err
}
