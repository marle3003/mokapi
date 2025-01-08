package dynamic

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/sortedmap"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
)

type PathResolver interface {
	Resolve(token string) (interface{}, error)
}

type Converter interface {
	ConvertTo(i interface{}) (interface{}, error)
}

func Resolve(ref string, element interface{}, config *Config, reader Reader) error {
	var err error

	if strings.HasPrefix(ref, "#") {
		err = resolveFragment(ref[1:], config, element)
	} else {
		var u *url.URL
		u, err = resolveUrl(ref, config)
		if err != nil {
			return err
		}

		var data interface{}
		if len(u.Fragment) > 0 && strings.Contains(u.Fragment, "/") && config.Data != nil {
			val := reflect.ValueOf(config.Data).Elem()
			data = reflect.New(val.Type()).Interface()
		} else {
			data = element
		}

		var sub *Config
		sub, err = reader.Read(removeFragment(u), data)
		if err != nil {
			return fmt.Errorf("resolve reference '%v' failed: %w", ref, err)
		}
		err = resolveFragment(u.Fragment, sub, element)
		AddRef(config, sub)
	}

	if err != nil {
		return fmt.Errorf("resolve reference '%v' failed: %w", ref, err)
	}
	return nil
}

func resolveFragment(fragment string, config *Config, resolved interface{}) (err error) {
	val := config.Data
	if fragment == "" {
		// resolve to current (root) element
	} else if strings.HasPrefix(fragment, "/") {
		val, err = resolvePath(fragment, config.Data)
	} else {
		val, err = config.Scope.Get(fragment)
	}
	if err != nil {
		return fmt.Errorf("resolve fragment '%v' failed: %w", fragment, err)
	}

	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		fRef := v.FieldByName("Ref")
		fValue := v.FieldByName("Value")
		if fRef.IsValid() && fValue.IsValid() {
			val = fValue.Interface()
		}
	}

	if val == nil {
		return fmt.Errorf("path '%v' not found", fragment)
	}

	if r, ok := val.(PathResolver); ok {
		if val, err = r.Resolve(""); err != nil {
			return
		}
	}

	vCursor := reflect.ValueOf(val)
	if reflect.Indirect(vCursor).Kind() == reflect.Map {
		reflect.Indirect(reflect.ValueOf(resolved)).Set(reflect.Indirect(vCursor))
		return
	}

	v2 := reflect.Indirect(reflect.ValueOf(resolved))
	if !vCursor.Type().AssignableTo(v2.Type()) && vCursor.Kind() == reflect.Ptr {
		if c, ok := val.(Converter); ok {
			if converted, err := c.ConvertTo(v2.Interface()); err == nil {
				vCursor = reflect.ValueOf(converted)
			}
		} else {
			vCursor = vCursor.Elem()
		}
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

	if m, ok := node.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
		if mv, ok := m.Get(token); ok {
			return mv, nil
		}
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
		for i := 0; i < rValue.NumField(); i++ {
			f := rValue.Field(i)
			if !f.CanInterface() {
				continue
			}

			json := rValue.Type().Field(i).Tag.Get("json")
			if strings.Split(json, ",")[0] == token {
				return f.Interface(), nil
			}
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

	return nil, fmt.Errorf("path element '%v' not found", token)
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

func resolvePath(path string, v interface{}) (interface{}, error) {
	tokens := strings.Split(path, "/")
	cursor := v
	var err error
	for _, t := range tokens[1:] {
		if r, ok := cursor.(PathResolver); ok {
			cursor, err = r.Resolve(t)
			if err != nil {
				return nil, err
			}
			continue
		}

		cursor, err = get(t, cursor)
		if err != nil {
			return nil, err
		}
	}

	return cursor, nil
}

func resolveUrl(ref string, cfg *Config) (*url.URL, error) {
	u, err := url.Parse(ref)
	if err != nil {
		return nil, err
	}

	if !u.IsAbs() {
		id := getId(cfg.Data)
		if id != "" {
			u, err = url.Parse(id)
			if err != nil {
				return nil, fmt.Errorf("parse URL from $id failed: %w", err)
			}
			return u.Parse(ref)
		} else {
			info := cfg.Info.Kernel()
			if len(info.Url.Opaque) > 0 {
				p := filepath.Join(filepath.Dir(info.Url.Opaque), u.Path)
				p = fmt.Sprintf("file:%v", p)
				if len(u.Fragment) > 0 {
					p = fmt.Sprintf("%v#%v", p, u.Fragment)
				}
				return url.Parse(p)
			} else {
				return info.Url.Parse(ref)
			}
		}
	}

	return u, err
}

func getId(v interface{}) string {
	if v == nil {
		return ""
	}

	val := reflect.Indirect(reflect.ValueOf(v))
	switch val.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			json := val.Type().Field(i).Tag.Get("json")
			if strings.Split(json, ",")[0] == "$id" {
				return f.Interface().(string)
			}
		}
	case reflect.Map:
		if a := val.MapIndex(reflect.ValueOf("$id")); a.IsValid() {
			return a.Interface().(string)
		}
	default:
		log.Warnf("resolve $id failed: type %v not supported", val.Kind())
	}

	return ""
}
