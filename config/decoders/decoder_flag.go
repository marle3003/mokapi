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

type context struct {
	path    string
	paths   []string
	element reflect.Value
	value   []string
}

func NewFlagDecoder() *FlagDecoder {
	return &FlagDecoder{fs: &file.Reader{}}
}

func NewFlagDecoderWithReader(reader file.FSReader) *FlagDecoder {
	return &FlagDecoder{fs: reader}
}

func (f *FlagDecoder) Decode(flags map[string][]string, element interface{}) error {
	keys := make([]string, 0, len(flags))
	for k := range flags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		paths := ParsePath(name)
		ctx := &context{path: name, paths: paths, value: flags[name], element: reflect.ValueOf(element)}
		err := f.setValue(ctx)
		if err != nil {
			return errors.Wrapf(err, "configuration error %v", name)
		}
	}

	return nil
}

func (f *FlagDecoder) setValue(ctx *context) error {
	switch ctx.element.Kind() {
	case reflect.Struct:
		if len(ctx.paths) == 0 {
			return f.convert(ctx.value[0], ctx.element)
		}
		err := ctx.setFieldFromStruct()
		if err != nil {
			return f.explode(ctx.element, ctx.paths[0], ctx.value)
		}
		return f.setValue(ctx)
	case reflect.Pointer:
		if ctx.element.IsNil() {
			ctx.element.Set(reflect.New(ctx.element.Type().Elem()))
		}
		ctx.element = ctx.element.Elem()
		return f.setValue(ctx)
	case reflect.String:
		s := strings.Trim(ctx.value[0], "\"")
		ctx.element.SetString(s)
		return nil
	case reflect.Int64:
		i, err := strconv.ParseInt(ctx.value[0], 10, 64)
		if err != nil {
			return fmt.Errorf("parse int64 failed: %v", err)
		}
		ctx.element.SetInt(i)
		return nil
	case reflect.Bool:
		b := false
		if ctx.value[0] == "" {
			b = true
		} else {
			var err error
			b, err = strconv.ParseBool(ctx.value[0])
			if err != nil {
				return fmt.Errorf("value %v cannot be parsed as bool: %v", ctx.value[0], err.Error())
			}
		}
		ctx.element.SetBool(b)
		return nil
	case reflect.Slice:
		return f.setArray(ctx)
	case reflect.Map:
		return f.setMap(ctx)
	case reflect.Interface:
		if len(ctx.value) == 0 || ctx.value[0] == "" {
			ctx.element.Set(reflect.ValueOf(true))
		} else {
			ctx.element.Set(reflect.ValueOf(ctx.value[0]))
		}
		return nil
	}

	return fmt.Errorf("unsupported config type: %v", ctx.element.Kind())
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

func (f *FlagDecoder) setArray(ctx *context) error {
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
		if len(ctx.value) == 1 {
			ctx.value = splitArrayItems(ctx.value[0])
		}

		if len(ctx.value) > 1 {
			// reset slice; remove default values
			ctx.element.Set(reflect.MakeSlice(ctx.element.Type(), 0, len(ctx.value)))
		}

		for _, v := range ctx.value {
			ptr := reflect.New(ctx.element.Type().Elem())
			ctxItem := &context{
				paths:   ctx.paths,
				element: ptr,
				value:   []string{v},
			}
			if err := f.setValue(ctxItem); err != nil {
				return err
			}
			ctx.element.Set(reflect.Append(ctx.element, ptr.Elem()))
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

func (f *FlagDecoder) setMap(ctx *context) error {
	m := ctx.element
	if m.IsNil() {
		m.Set(reflect.MakeMap(ctx.element.Type()))
	}

	key := reflect.ValueOf(ctx.paths[0])
	var ptr reflect.Value
	ptr = reflect.New(reflect.PointerTo(ctx.element.Type().Elem()))

	if m.MapIndex(key).IsValid() {
		ptr.Elem().Set(reflect.New(ctx.element.Type().Elem()))
		ptr.Elem().Elem().Set(ctx.element.MapIndex(key))
	}
	if err := f.setValue(ctx.Next(ptr)); err != nil {
		return err
	}

	m.SetMapIndex(key, ptr.Elem().Elem())

	return nil
}

func (f *FlagDecoder) explode(v reflect.Value, name string, value []string) error {
	field := getFieldByTag(v, name, "explode")
	if !field.IsValid() {
		return fmt.Errorf("not found")
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

func getFieldByTag(v reflect.Value, name, tag string) reflect.Value {
	for i := 0; i < v.NumField(); i++ {
		explode := v.Type().Field(i).Tag.Get(tag)
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
			}
			if len(u.Path) > 0 {
				path += u.Path
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
			err = f.setValue(&context{paths: []string{kv[0]}, value: []string{kv[1]}, element: v})
			if err != nil {
				return err
			}
		}
		return nil
	} else if kind == reflect.Slice {
		return f.setValue(&context{paths: []string{}, element: v, value: []string{s}})
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

	return f.setJson(v, m)
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

func (f *FlagDecoder) setJson(element reflect.Value, i interface{}) error {
	switch o := i.(type) {
	case int64, string, bool:
		element.Set(reflect.ValueOf(i))
	case []interface{}:
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
					return fmt.Errorf("configuration not found")
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

func invertFlag(value string) (string, error) {
	flag := false
	if value != "" {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return "", fmt.Errorf("value %v cannot be parsed as bool: %v", value, err.Error())
		}
		flag = !b
	}
	return fmt.Sprintf("%v", flag), nil
}

func (c *context) setFieldFromStruct() error {
	name := strings.ToLower(c.paths[0])
	field := c.element.FieldByNameFunc(func(f string) bool {
		return strings.ToLower(f) == name
	})
	if field.IsValid() {
		c.Next(field)
		return nil
	}

	if c.paths[0] == "no" && len(c.paths) == 2 {
		name = strings.ToLower(c.paths[1])
		field = c.element.FieldByNameFunc(func(f string) bool {
			return strings.ToLower(f) == c.paths[1]
		})
		if field.IsValid() {
			value, err := invertFlag(c.value[0])
			if err != nil {
				return err
			}
			c.value = []string{value}
			c.Next(field)
			return nil
		}
	}

	for i := 0; i < c.element.NumField(); i++ {
		f := c.element.Type().Field(i)

		tag := f.Tag.Get("flag")
		if len(tag) == 0 {
			tag = f.Tag.Get("name")
			if len(tag) == 0 {
				continue
			}
		}
		name = ""
		for j := 0; j < len(c.paths); j++ {
			if len(name) > 0 {
				name += "-"
			}
			name += c.paths[j]
			if tag == name {
				c.element = c.element.Field(i)
				c.paths = c.paths[j+1:]
				return nil
			}
		}
	}
	return fmt.Errorf("no configuration found")
}

func (c *context) Next(element reflect.Value) *context {
	c.paths = c.paths[1:]
	c.element = element
	return c
}
