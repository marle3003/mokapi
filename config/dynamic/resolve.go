package dynamic

import (
	"fmt"
	"mokapi/sortedmap"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
)

type PathResolver interface {
	Resolve(token string) (interface{}, error)
}

type Converter interface {
	ConvertTo(i any) (any, error)
}

type FromConverter interface {
	ConvertFrom(i any) (any, error)
}

func resolve[T any](ref string, config *Config, reader Reader) (T, error) {
	var err error
	var result T

	var fragment string
	isLocal := true
	parent := config
	if len(ref) > 0 && !strings.HasPrefix(ref, "#") {
		fragment, config, err = resolveResource[T](ref, config, reader)
		if err != nil {
			return result, fmt.Errorf("resolve reference '%v' failed: %w", ref, err)
		}
		isLocal = false
	} else if len(ref) > 0 {
		fragment = ref[1:]
	}

	result, err = resolveFragment[T](fragment, config, false)

	if err != nil {
		return result, fmt.Errorf("resolve reference '%v' failed: %w", ref, err)
	}

	// Parse the referenced schema again in the current context.
	// This ensures nested $ref and $dynamicRef are resolved relative
	// to the correct dynamic scope.
	v := reflect.ValueOf(result)
	p, ok := v.Interface().(Parser)
	if ok {
		if !isLocal {
			// set parent scope hierarchy
			config = &Config{Raw: config.Raw, Data: copyData(config.Data), Info: config.Info}
			config.Scope.SetParent(parent.Scope)
		}
		if !config.EnterRef(ref) {
			return result, nil
		}
		defer config.LeaveRef(ref)

		err = p.Parse(config, reader)
		if err != nil {
			return result, fmt.Errorf("resolve reference '%v' failed: %w", ref, err)
		}
	}

	return result, nil
}

func ResolveDynamic[T any](ref string, config *Config, reader Reader) (T, error) {
	var err error
	var result T

	fragment := ref[1:]
	if !strings.HasPrefix(ref, "#") {
		fragment, config, err = resolveResource[T](ref, config, reader)
		if err != nil {
			return result, fmt.Errorf("resolve reference '%v' failed: %w", ref, err)
		}
	}

	result, err = resolveFragment[T](fragment, config, true)

	if err != nil {
		return result, fmt.Errorf("resolve reference '%v' failed: %w", ref, err)
	}
	return result, nil
}

func resolveFragment[T any](fragment string, config *Config, dynamic bool) (result T, err error) {
	val := config.Data
	if fragment == "" {
		// resolve to current (root) element
	} else if strings.HasPrefix(fragment, "/") {
		val, err = resolvePath(fragment, config.Data)
	} else if dynamic {
		val, err = config.Scope.GetDynamic(fragment)
	} else {
		val, err = config.Scope.GetLexical(fragment)
	}
	if err != nil {
		return
	}

	result, err = convertTo[T](val)
	return
}

func convertTo[T any](val any) (T, error) {
	if val == nil {
		return *new(T), fmt.Errorf("value is null")
	}

	if p, ok := val.(PathResolver); ok {
		var err error
		val, err = p.Resolve("")
		if err != nil {
			return *new(T), err
		}
	}

	val = convert[T](val)

	// types are identical
	if v, ok := val.(T); ok {
		return v, nil
	}

	valType := reflect.TypeOf(val)
	targetType := reflect.TypeOf((*T)(nil)).Elem()

	// val is pointer but T not
	if valType != nil && valType.Kind() == reflect.Ptr && valType.Elem() == targetType {
		v := reflect.ValueOf(val)
		if !v.IsNil() {
			return v.Elem().Interface().(T), nil
		}
	}

	// T is pointer but val not
	if valType != nil && reflect.PointerTo(valType) == targetType {
		vp := reflect.New(valType)
		vp.Elem().Set(reflect.ValueOf(val))
		return vp.Interface().(T), nil
	}

	var result T
	return result, fmt.Errorf("expected type %T, got %T", result, val)
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
		t = strings.ReplaceAll(t, "~1", "/")
		t = strings.ReplaceAll(t, "~0", "~")
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

	if u.IsAbs() {
		return u, nil
	}

	id := getId(cfg.Data)
	if id != "" {
		u, err = url.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("parse URL from $id failed: %w", err)
		}
		if u.IsAbs() {
			return u.Parse(ref)
		}
	}

	info := cfg.Info.Kernel()
	if info.Url == nil {
		return u, nil
	}

	if len(info.Url.Opaque) > 0 {
		p := filepath.Join(filepath.Dir(info.Url.Opaque), u.Path)
		p = fmt.Sprintf("file:%v", p)
		if len(u.Fragment) > 0 {
			p = fmt.Sprintf("%v#%v", p, u.Fragment)
		}
		return url.Parse(p)
	}

	refURL := info.Url.ResolveReference(u)
	if u.Fragment != "" {
		refURL.Fragment = u.Fragment
	}
	return refURL, nil
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

func resolveResource[T any](ref string, config *Config, reader Reader) (string, *Config, error) {
	u, err := resolveUrl(ref, config)
	if err != nil {
		return "", nil, err
	}

	var data interface{}
	if len(u.Fragment) > 0 && strings.Contains(u.Fragment, "/") && config.Data != nil {
		val := reflect.ValueOf(config.Data).Elem()
		data = reflect.New(val.Type()).Interface()
	} else {
		var result T
		data = result
	}

	sub, err := reader.Read(removeFragment(u), data)
	if err == nil {
		AddRef(config, sub)
	}
	return u.Fragment, sub, err
}

func copyData(input interface{}) interface{} {
	val := reflect.ValueOf(input)

	if val.Kind() != reflect.Ptr {
		// Not a pointer — return the original or handle differently
		return input
	}

	// Create a new object of the same type
	c := reflect.New(val.Elem().Type())

	// Set the value of the new object to the original
	c.Elem().Set(val.Elem())

	return c.Interface()
}

func convert[T any](val any) any {
	var target T

	if c, ok := val.(Converter); ok {
		result, err := c.ConvertTo(target)
		if err == nil {
			return result
		}
	}

	v := reflect.ValueOf(target)
	if !v.IsValid() || !v.CanInterface() {
		return val
	}
	c, ok := v.Interface().(FromConverter)
	if !ok {
		return val
	}
	result, err := c.ConvertFrom(val)
	if err == nil {
		return result
	}
	return val
}
