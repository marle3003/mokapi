package dynamic

import (
	"fmt"
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

	if len(u.Path) > 0 || len(u.Host) > 0 {
		if !u.IsAbs() {
			info := config.Info.Kernel()
			if len(info.Url.Opaque) > 0 {
				p := filepath.Join(filepath.Dir(info.Url.Opaque), u.Path)
				p = fmt.Sprintf("file:%v", p)
				if len(u.Fragment) > 0 {
					p = fmt.Sprintf("%v#%v", p, u.Fragment)
				}
				u, err = url.Parse(p)
			} else {
				u, err = info.Url.Parse(ref)
			}
		}

		var data interface{}
		if len(u.Fragment) > 0 {
			val := reflect.ValueOf(config.Data).Elem()
			data = reflect.New(val.Type()).Interface()
		} else {
			data = element
		}

		f, err := reader.Read(removeFragment(u), data)
		if err != nil {
			return fmt.Errorf("resolve reference '%v' failed: %w", ref, err)
		}

		err = resolvePath(u.Fragment, f.Data, element)
		if err != nil {
			return fmt.Errorf("resolve reference '%v' failed: %w", ref, err)
		}
		AddRef(config, f)
		return nil
	}

	err = resolvePath(u.Fragment, config.Data, element)
	if err != nil {
		return fmt.Errorf("resolve reference '%v' failed: %w", ref, err)
	}
	return nil
}

func resolvePath(path string, cursor interface{}, resolved interface{}) (err error) {
	tokens := strings.Split(path, "/")

	for i, t := range tokens[1:] {
		if r, ok := cursor.(PathResolver); ok {
			cursor, err = r.Resolve(strings.Join(tokens[i+1:], "/"))
			if err != nil {
				return err
			}
			break
		}

		cursor, err = get(t, cursor)
		if err != nil {
			return err
		}
	}

	v := reflect.ValueOf(cursor)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		fRef := v.FieldByName("Ref")
		fValue := v.FieldByName("Value")
		if fRef.IsValid() && fValue.IsValid() {
			cursor = fValue.Interface()
		}
	}

	if cursor == nil {
		return fmt.Errorf("path '%v' not found", path)
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
	if !vCursor.Type().AssignableTo(v2.Type()) && vCursor.Kind() == reflect.Ptr {
		vCursor = vCursor.Elem()
	}

	if !vCursor.Type().AssignableTo(v2.Type()) {
		return fmt.Errorf("expected type %v, got %v", v2.Type(), vCursor.Type())
	}

	v2.Set(vCursor)

	return
}

func get(token string, node interface{}) (interface{}, error) {
	if len(token) == 0 {
		return node, nil
	}

	rValue := reflect.Indirect(reflect.ValueOf(node))
	switch rValue.Kind() {
	case reflect.Struct:
		// if node is a "ref wrapper" like SchemaRef
		if f := rValue.FieldByName("Value"); f.IsValid() {
			return get(token, f.Interface())
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
	default:
		break
	}

	return nil, fmt.Errorf("invalid token reference %q", token)
}

func caseInsensitiveFieldByName(v reflect.Value, name string) reflect.Value {
	name = strings.ToLower(name)
	return v.FieldByNameFunc(func(n string) bool { return strings.ToLower(n) == name })
}

func removeFragment(u *url.URL) *url.URL {
	c := new(url.URL)
	// shallow copy
	*c = *u
	c.Fragment = ""
	return c
}
