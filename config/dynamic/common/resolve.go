package common

import (
	"fmt"
	"github.com/pkg/errors"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
)

type PathResolver interface {
	Resolve(token string) (interface{}, error)
}

func Resolve(ref string, element interface{}, config *Config, reader Reader) error {
	u, err := url.Parse(ref)
	if err != nil {
		return err
	}

	if len(u.Path) > 0 {
		if !u.IsAbs() {
			if len(config.Url.Opaque) > 0 {
				p := filepath.Join(filepath.Dir(config.Url.Opaque), u.Path)
				p = fmt.Sprintf("file:%v", p)
				if len(u.Fragment) > 0 {
					p = fmt.Sprintf("%v#%v", p, u.Fragment)
				}
				u, err = url.Parse(p)
			} else {
				u, err = config.Url.Parse(ref)
			}
		}

		opts := []ConfigOptions{WithParent(config)}
		if len(u.Fragment) > 0 {
			val := reflect.ValueOf(config.Data).Elem()
			opts = append(opts, WithData(reflect.New(val.Type()).Interface()))
		} else {
			opts = append(opts, WithData(element))
		}

		f, err := reader.Read(u, opts...)
		if err != nil {
			return fmt.Errorf("unable to read %v: %v", u, err)
		}
		err = ResolvePath(u.Fragment, f.Data, element)
		if err != nil {
			return errors.Wrapf(err, "unable to resolve reference %v", ref)
		}
		return nil
	}

	return ResolvePath(u.Fragment, config.Data, element)
}

func ResolvePath(path string, cursor interface{}, resolved interface{}) (err error) {
	tokens := strings.Split(path, "/")

	for _, t := range tokens[1:] {
		cursor, err = Get(t, cursor)
		if err != nil {
			return
		}
	}

	if cursor == nil {
		return fmt.Errorf("unresolved path: %q", path)
	}

	if r, ok := cursor.(PathResolver); ok {
		if cursor, err = r.Resolve(""); err != nil {
			return
		}
	}

	vCursor := reflect.ValueOf(cursor)
	if reflect.Indirect(vCursor).Kind() == reflect.Map {
		reflect.Indirect(reflect.ValueOf(resolved)).Set(reflect.Indirect(vCursor))
		return
	}

	v2 := reflect.Indirect(reflect.ValueOf(resolved))
	if !vCursor.Type().AssignableTo(v2.Type()) {
		vCursor = vCursor.Elem()
	}

	if !vCursor.Type().AssignableTo(v2.Type()) {
		return fmt.Errorf("expected type %v, got %v", v2.Type(), vCursor.Type())
	}

	v2.Set(vCursor)

	return
}

func Get(token string, node interface{}) (interface{}, error) {
	if len(token) == 0 {
		return node, nil
	}

	rValue := reflect.Indirect(reflect.ValueOf(node))

	if r, ok := node.(PathResolver); ok {
		return r.Resolve(token)
	}

	switch rValue.Kind() {
	case reflect.Struct:
		// if node is a "ref wrapper" like SchemaRef
		if f := rValue.FieldByName("Value"); f.IsValid() {
			return Get(token, f.Interface())
		}
		if f := caseInsensitiveFieldByName(rValue, token); f.IsValid() {
			return f.Interface(), nil
		}
	case reflect.Map:
		mv := rValue.MapIndex(reflect.ValueOf(token))
		if mv.IsValid() {
			v := reflect.Indirect(mv)
			// if map value is a "ref wrapper" like SchemaRef
			if v.Kind() == reflect.Struct {
				if f := v.FieldByName("Value"); f.IsValid() {
					return f.Interface(), nil
				}
			}
			return mv.Interface(), nil
		}
	}

	return nil, fmt.Errorf("invalid token reference %q", token)
}

func caseInsensitiveFieldByName(v reflect.Value, name string) reflect.Value {
	name = strings.ToLower(name)
	return v.FieldByNameFunc(func(n string) bool { return strings.ToLower(n) == name })
}
