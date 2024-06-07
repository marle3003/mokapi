package decoders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"mokapi/config/dynamic/provider/file"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type FlagDecoder struct {
	fs file.FSReader
}

func NewFlagDecoder() *FlagDecoder {
	return &FlagDecoder{fs: &file.Reader{}}
}

func (f *FlagDecoder) Decode(flags map[string][]string, element interface{}) error {
	keys := make([]string, 0, len(flags))
	for k := range flags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		paths := f.parsePath(name)
		for _, value := range flags[name] {
			err := f.setValue(paths, value, reflect.ValueOf(element))
			if err != nil {
				return errors.Wrapf(err, "configuration error %v", name)
			}
		}
	}

	return nil
}

func (f *FlagDecoder) setValue(paths []string, value string, element reflect.Value) error {
	switch element.Kind() {
	case reflect.Struct:
		if len(paths) == 0 {
			return f.convert(value, element)
		}
		name := strings.ToLower(paths[0])
		field := element.FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == name })
		if !field.IsValid() {
			if strings.HasPrefix(paths[0], "no-") {
				return invertFlag(paths[0], value, element)
			} else {
				return f.explode(element, paths[0], value)
			}
		}

		return f.setValue(paths[1:], value, field)
	case reflect.Pointer:
		if element.IsNil() {
			element.Set(reflect.New(element.Type().Elem()))
		}
		return f.setValue(paths, value, element.Elem())
	case reflect.String:
		s := strings.Trim(value, "\"")
		element.SetString(s)
		return nil
	case reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("parse int64 failed: %v", err)
		}
		element.SetInt(i)
		return nil
	case reflect.Bool:
		b := false
		if value == "" {
			b = true
		} else {
			var err error
			b, err = strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("value %v cannot be parsed as bool: %v", value, err.Error())
			}
		}
		element.SetBool(b)
		return nil
	case reflect.Slice:
		return f.setArray(paths, value, element)
	case reflect.Map:
		return f.setMap(paths, value, element)
	}

	panic(fmt.Errorf("unsupported config type: %v", element.Kind()))
}

func (f *FlagDecoder) parsePath(key string) []string {
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

func (f *FlagDecoder) setArray(paths []string, value string, element reflect.Value) error {
	if len(paths) > 0 {
		index, err := f.parseArrayIndex(paths[0])
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

		return f.setValue(paths[1:], value, element.Index(index))
	} else {
		values := splitArrayItems(value)
		for _, v := range values {
			ptr := reflect.New(element.Type().Elem())
			if err := f.setValue(paths, v, ptr); err != nil {
				return err
			}
			element.Set(reflect.Append(element, ptr.Elem()))
		}
	}

	return nil
}

func (f *FlagDecoder) parseArrayIndex(path string) (int, error) {
	if strings.HasPrefix(path, "[") {
		s := strings.TrimPrefix(path, "[")
		s = strings.TrimSuffix(s, "]")
		return strconv.Atoi(s)

	}
	return strconv.Atoi(path)
}

func (f *FlagDecoder) setMap(paths []string, value string, element reflect.Value) error {
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
	if err := f.setValue(paths[1:], value, ptr); err != nil {
		return err
	}

	element.SetMapIndex(key, ptr.Elem().Elem())

	return nil
}

func (f *FlagDecoder) explode(v reflect.Value, name string, value string) error {
	field := f.getFieldByTag(v, name)
	if !field.IsValid() {
		return fmt.Errorf("not found")
	}

	o := reflect.New(field.Type().Elem())
	err := f.convert(value, o.Elem())
	if err != nil {
		return err
	}
	field.Set(reflect.Append(field, o.Elem()))
	return nil
}

func (f *FlagDecoder) getFieldByTag(v reflect.Value, name string) reflect.Value {
	for i := 0; i < v.NumField(); i++ {
		explode := v.Type().Field(i).Tag.Get("explode")
		if explode == name {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}

func (f *FlagDecoder) convert(s string, v reflect.Value) error {
	u, err := url.ParseRequestURI(s)
	if err == nil {
		switch u.Scheme {
		case "file":
			var path string
			if len(u.Host) > 0 {
				path = u.Host
			} else if len(u.Path) > 0 {
				path = u.Path
			}
			if len(u.Opaque) > 0 {
				path = u.Opaque
			}
			b, err := f.fs.ReadFile(path)
			if err != nil {
				return err
			}
			// remove bom sequence if present
			if len(b) >= 4 && bytes.Equal(b[0:3], file.Bom) {
				b = b[3:]
			}
			s = string(b)
		}
	}

	kind := v.Type().Kind()
	if kind == reflect.Struct {
		err = f.convertJson(s, v)
		if err == nil {
			return nil
		}
	}

	if kind == reflect.Struct {
		pairs := strings.Split(s, ",")
		for _, pair := range pairs {
			kv := strings.Split(pair, "=")
			if len(kv) != 2 {
				return fmt.Errorf("parse shorthand failed: %v", s)
			}
			err = f.setValue([]string{kv[0]}, kv[1], v)
			if err != nil {
				return err
			}
		}
		return nil
	} else if kind == reflect.Slice {
		return f.setValue([]string{}, s, v)
	} else if kind == reflect.String {
		v.Set(reflect.ValueOf(s))
		return nil
	}

	return fmt.Errorf("not supported")
}

func (f *FlagDecoder) convertJson(s string, v reflect.Value) error {
	m := map[string]interface{}{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return err
	}

	return f.set(v, m)
}

func splitArrayItems(s string) []string {
	quoted := false
	return strings.FieldsFunc(s, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})
}

func (f *FlagDecoder) set(element reflect.Value, i interface{}) error {
	switch o := i.(type) {
	case int64, string, bool:
		element.Set(reflect.ValueOf(i))
	case []interface{}:
		for _, item := range o {
			ptr := reflect.New(element.Type().Elem())
			err := f.set(ptr.Elem(), item)
			if err != nil {
				return err
			}
			element.Set(reflect.Append(element, ptr.Elem()))
		}
	case map[string]interface{}:
		for k, v := range o {
			field := element.FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == strings.ToLower(k) })
			if field.IsValid() {
				err := f.set(field, v)
				if err != nil {
					return err
				}
			} else {
				field = f.getFieldByTag(element, k)
				if !field.IsValid() {
					return fmt.Errorf("configuration not found")
				}
				ptr := reflect.New(field.Type().Elem())
				err := f.set(ptr.Elem(), v)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(field, ptr.Elem()))
			}
		}
	}

	return nil
}

func invertFlag(name string, value string, element reflect.Value) error {
	name = strings.ToLower(strings.TrimPrefix(name, "no-"))
	field := element.FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == name })
	if !field.IsValid() {
		return fmt.Errorf("configuration not found")
	}
	flag := false
	if value != "" {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("value %v cannot be parsed as bool: %v", value, err.Error())
		}
		flag = !b
	}
	field.Set(reflect.ValueOf(flag))
	return nil
}
