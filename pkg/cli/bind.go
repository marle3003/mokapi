package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic/provider/file"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type flagConfigBinder struct{}

type bindContext struct {
	path    string
	paths   []string
	element reflect.Value
	value   any
}

func (f *flagConfigBinder) Decode(flags *FlagSet, element interface{}) error {
	return flags.Visit(func(flag *Flag) error {
		paths := ParsePath(flag.Name)
		v := flag.Value.Value()
		ctx := &bindContext{path: flag.Name, paths: paths, value: v, element: reflect.ValueOf(element)}
		err := f.setValue(ctx)
		if err != nil {
			return fmt.Errorf("configuration error '%v' value '%v': %w", flag.Name, v, err)
		}
		return nil
	})
}

func (f *flagConfigBinder) setValue(ctx *bindContext) error {
	switch ctx.element.Kind() {
	case reflect.Struct:
		if len(ctx.paths) == 0 {
			return f.convert(ctx.value, ctx.element)
		}
		err := ctx.setFieldFromStruct()
		if err != nil {
			if len(ctx.paths) == 1 {
				if arr, ok := ctx.value.([]string); ok {
					return f.explode(ctx.element, ctx.paths[0], arr)
				}
				if s, ok := ctx.value.(string); ok {
					return f.explode(ctx.element, ctx.paths[0], []string{s})
				}
			}
			// skip: field not found
			return nil
		}
		return f.setValue(ctx)
	case reflect.Pointer:
		if ctx.element.IsNil() {
			ctx.element.Set(reflect.New(ctx.element.Type().Elem()))
		}
		ctx.element = ctx.element.Elem()
		return f.setValue(ctx)
	case reflect.Slice:
		return f.setArray(ctx)
	case reflect.Map:
		return f.setMap(ctx)
	case reflect.Bool:
		switch v := ctx.value.(type) {
		case bool:
			ctx.element.SetBool(v)
			return nil
		case string:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("value %v cannot be parsed as bool: %w", ctx.value, err)
			}
			ctx.element.SetBool(b)
			return nil
		}
		return fmt.Errorf("value %v cannot be parsed as bool", ctx.value)
	case reflect.String:
		if s, ok := ctx.value.(string); ok {
			return f.convert(s, ctx.element)
		}
		return fmt.Errorf("expected string but got '%v'", ctx.value)
	case reflect.Int:
		if i, ok := ctx.value.(int); ok {
			ctx.element.SetInt(int64(i))
			return nil
		}
		return fmt.Errorf("expected integer but got '%v'", ctx.value)
	case reflect.Int64:
		if i, ok := ctx.value.(int); ok {
			ctx.element.SetInt(int64(i))
			return nil
		}
		if s, ok := ctx.value.(string); ok {
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return fmt.Errorf("parse int64 failed: %v", err)
			}
			ctx.element.SetInt(i)
			return nil
		}
		return fmt.Errorf("value %v cannot be parsed as integer", ctx.value)
	case reflect.Interface:
		if ctx.value == "" {
			ctx.element.Set(reflect.ValueOf(true))
		} else {
			ctx.element.Set(reflect.ValueOf(ctx.value))
		}
		return nil
	default:
		return fmt.Errorf("unsupported config type: %v", ctx.element.Kind())
	}
}

func ParsePath(key string) []string {
	var paths []string
	split := strings.FieldsFunc(key, func(r rune) bool {
		return r == '.' || r == '-'
	})

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

func (f *flagConfigBinder) setArray(ctx *bindContext) error {
	if len(ctx.paths) > 0 {
		index, err := f.parseArrayIndex(ctx.paths[0])
		if err != nil {
			return fmt.Errorf("parse array index failed: %v", err)
		}

		if index >= ctx.element.Cap() {
			n := index + 1
			nCap := 2 * n
			if nCap < 4 {
				nCap = 4
			}
			if ctx.element.IsNil() {
				s := reflect.MakeSlice(ctx.element.Type(), n, nCap)
				ctx.element.Set(s)
			} else {
				s := reflect.MakeSlice(ctx.element.Type(), n, nCap)
				reflect.Copy(s, ctx.element)
				ctx.element.Set(s)
			}
		}
		if index >= ctx.element.Len() {
			ctx.element.SetLen(index + 1)
		}

		return f.setValue(ctx.Next(ctx.element.Index(index)))
	} else {
		var values []string
		if arr, ok := ctx.value.([]string); ok {
			if arr == nil {
				return nil
			}
			values = arr
		} else if s, ok := ctx.value.(string); ok {
			values = []string{s}
		}

		if len(values) == 1 {
			values = splitArrayItems(values[0])
		}

		if len(values) > 0 {
			// reset slice; remove default values
			ctx.element.Set(reflect.MakeSlice(ctx.element.Type(), 0, len(values)))
		}

		for _, v := range values {
			ptr := reflect.New(ctx.element.Type().Elem())
			ctxItem := &bindContext{
				paths:   ctx.paths,
				element: ptr,
				value:   v,
			}
			if err := f.setValue(ctxItem); err != nil {
				return err
			}
			ctx.element.Set(reflect.Append(ctx.element, ptr.Elem()))
		}
	}

	return nil
}

func (f *flagConfigBinder) parseArrayIndex(path string) (int, error) {
	if strings.HasPrefix(path, "[") {
		s := strings.TrimPrefix(path, "[")
		s = strings.TrimSuffix(s, "]")
		return strconv.Atoi(s)

	}
	return strconv.Atoi(path)
}

func (f *flagConfigBinder) setMap(ctx *bindContext) error {
	var values []string
	if arr, ok := ctx.value.([]string); ok {
		values = arr
	} else if s, ok := ctx.value.(string); ok {
		if s == "" {
			return nil
		}
		values = []string{s}
	}

	m := ctx.element
	if m.IsNil() {
		m.Set(reflect.MakeMap(ctx.element.Type()))
	}

	var key reflect.Value
	if len(ctx.paths) >= 1 {
		key = reflect.ValueOf(ctx.paths[0])
	} else if len(ctx.paths) == 0 {
		if len(values) == 1 {
			kv := strings.Split(values[0], "=")
			if len(kv) != 2 {
				return fmt.Errorf("expected value with key value pair for map like key=value: %s", values[0])
			}
			key = reflect.ValueOf(kv[0])
			ctx.value = kv[1]
		}
	}

	if !key.IsValid() {
		return fmt.Errorf("expected key to set map value")
	}

	v := m.MapIndex(key)
	if !v.IsValid() {
		v = reflect.New(m.Type().Elem())
	} else {
		p := reflect.New(m.Type().Elem())
		p.Elem().Set(v)
		v = p
	}

	ctx.element = v

	if len(ctx.paths) >= 1 {
		v = ctx.element
		if err := f.setValue(ctx.Next(ctx.element)); err != nil {
			return err
		}
	} else {
		if err := f.setValue(ctx); err != nil {
			return err
		}
	}

	m.SetMapIndex(key, v.Elem())

	return nil
}

func (f *flagConfigBinder) explode(v reflect.Value, name string, value []string) error {
	field := getFieldByTag(v, name, "explode")
	if !field.IsValid() {
		return nil
	}

	for _, val := range value {
		o := reflect.New(field.Type().Elem())
		err := f.convert(val, o.Elem())
		if err != nil {
			return err
		}
		field.Set(reflect.Append(field, o.Elem()))
	}

	return nil
}

func getFieldByTag(structValue reflect.Value, name, tag string) reflect.Value {
	for i := 0; i < structValue.NumField(); i++ {
		v := structValue.Type().Field(i).Tag.Get(tag)
		tagValues := strings.Split(v, ",")
		for _, tagValue := range tagValues {
			if tagValue == name {
				return structValue.Field(i)
			}
		}
	}
	return reflect.Value{}
}

func (f *flagConfigBinder) convert(value any, target reflect.Value) error {
	kind := target.Type().Kind()

	if s, ok := value.(string); ok {
		u, err := url.ParseRequestURI(s)
		if err == nil {
			switch u.Scheme {
			case "file":
				var path string
				if len(u.Host) > 0 {
					path = u.Host
				}
				if len(u.Path) > 0 {
					path += u.Path
				}
				if len(u.Opaque) > 0 {
					path = u.Opaque
				}
				b, err := fileReader.ReadFile(path)
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

		if kind == reflect.Struct {
			err = f.convertJson(s, target)
			if err == nil {
				return nil
			}
		}

		if kind == reflect.Struct {
			if s == "" {
				return nil
			}
			pairs := strings.Split(s, ",")
			for _, pair := range pairs {
				kv := strings.Split(pair, "=")
				if len(kv) != 2 {
					return fmt.Errorf("parse shorthand failed: %v", s)
				}
				err = f.setValue(&bindContext{paths: []string{kv[0]}, value: kv[1], element: target})
				if err != nil {
					return err
				}
			}
			return nil
		} else if kind == reflect.Slice {
			return f.setValue(&bindContext{paths: []string{}, element: target, value: []string{s}})
		} else if kind == reflect.String {
			v := reflect.ValueOf(s)
			t := target.Type()
			if v.Type().AssignableTo(t) {
				target.Set(v)
				return nil
			} else if v.Type().ConvertibleTo(t) {
				target.Set(v.Convert(t))
				return nil
			}
		}

		value = s
	}

	v := reflect.ValueOf(value)
	if v.Type().AssignableTo(target.Type()) {
		target.Set(v)
		return nil
	}

	return fmt.Errorf("not supported")
}

func (f *flagConfigBinder) convertJson(s string, v reflect.Value) error {
	m := map[string]interface{}{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return err
	}

	return f.setJson(v, m)
}

func splitArrayItems(s string) []string {
	quoted := false
	splitC := getSplitCharsForList(s)
	items := strings.FieldsFunc(s, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && slices.Contains(splitC, r)
	})
	for i, item := range items {
		items[i] = strings.Trim(item, "\"")
	}
	return items
}

func (f *flagConfigBinder) setJson(element reflect.Value, i interface{}) error {
	switch o := i.(type) {
	case float64:
		// currently, config uses only int64 as number
		i = int64(o)
		element.Set(reflect.ValueOf(i))
	case int64, string, bool:
		element.Set(reflect.ValueOf(i))
	case []interface{}:
		// reset array
		element.Set(reflect.MakeSlice(element.Type(), 0, len(o)))
		for _, item := range o {
			ptr := reflect.New(element.Type().Elem())
			err := f.setJson(ptr.Elem(), item)
			if err != nil {
				return err
			}
			element.Set(reflect.Append(element, ptr.Elem()))
		}
	case map[string]interface{}:
		for k, v := range o {
			field := element.FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == strings.ToLower(k) })
			if field.IsValid() {
				err := f.setJson(field, v)
				if err != nil {
					return err
				}
			} else {
				field = getFieldByTag(element, k, "explode")
				if !field.IsValid() {
					return nil
				}
				ptr := reflect.New(field.Type().Elem())
				err := f.setJson(ptr.Elem(), v)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(field, ptr.Elem()))
			}
		}
	}

	return nil
}

func (c *bindContext) setFieldFromStruct() error {
	name := strings.ToLower(c.paths[0])
	field := c.element.FieldByNameFunc(func(f string) bool {
		return strings.ToLower(f) == name
	})
	if field.IsValid() {
		c.Next(field)
		return nil
	}

	for i := 0; i < c.element.NumField(); i++ {
		f := c.element.Type().Field(i)

		var names []string
		tag := f.Tag.Get("flag")
		if len(tag) == 0 {
			tag = f.Tag.Get("name")
			if len(tag) > 0 {
				names = append(names, tag)
			} else {
				tag = f.Tag.Get("aliases")
				if len(tag) > 0 {
					names = append(names, strings.Split(tag, ",")...)
				} else {
					continue
				}
			}
		} else {
			names = append(names, tag)
		}

		name = ""
		for j := 0; j < len(c.paths); j++ {
			if len(name) > 0 {
				name += "-"
			}
			name += c.paths[j]
			if slices.Contains(names, name) {
				c.element = c.element.Field(i)
				c.paths = c.paths[j+1:]
				return nil
			}
		}
	}
	return fmt.Errorf("no configuration found")
}

func (c *bindContext) Next(element reflect.Value) *bindContext {
	c.paths = c.paths[1:]
	c.element = element
	return c
}

func getSplitCharsForList(s string) []rune {
	s = strings.Trim(s, "")
	if strings.Contains(s, ",") && strings.Contains(s, " ") || isJsonValue(s) {
		return []rune{' '}
	}
	return []rune{' ', ','}
}

func isJsonValue(s string) bool {
	return (strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")) ||
		(strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}"))
}
