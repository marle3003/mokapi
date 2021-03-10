package asyncApi

import (
	"fmt"
	"mokapi/config/dynamic"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
)

type refResolver struct {
	reader dynamic.ConfigReader
	path   string
	config *Config
	eh     dynamic.ChangeEventHandler
}

func (r refResolver) resolveConfig() error {
	if err := r.resolveMokapiRef(r.config.Info.Mokapi); err != nil {
		return err
	}

	return nil
}

func (r refResolver) resolveMokapiRef(m *MokapiRef) error {
	if m == nil {
		return nil
	}

	if len(m.Ref) > 0 && m.Value == nil {
		u, err := url.Parse(m.Ref)
		if err != nil {
			return err
		}

		if !isLocalRef(m.Ref) {
			err := r.loadFrom(u.Path, &m.Value)
			if err != nil {
				return err
			}
		} else {
			err := r.resolve(u.Fragment, r.config, &m.Value)
			if err != nil {
				return err
			}
		}
	}

	if m.Value == nil {
		return nil
	}

	return nil
}

func (r refResolver) resolve(ref string, node interface{}, val interface{}) (err error) {
	tokens := strings.Split(ref[1:], "/")

	i := node
	for _, t := range tokens {
		i, err = get(t, i)
	}

	if i == nil {
		return fmt.Errorf("found unresolved ref: %q", ref)
	}

	reflect.ValueOf(val).Elem().Set(reflect.ValueOf(i))

	return
}

func get(token string, node interface{}) (interface{}, error) {
	rValue := reflect.Indirect(reflect.ValueOf(node))

	switch rValue.Kind() {
	case reflect.Struct:
		f := rValue.FieldByName(token)
		if f.IsValid() {
			return f.Interface(), nil
		}
	case reflect.Map:
		mv := rValue.MapIndex(reflect.ValueOf(token))
		if mv.IsValid() {
			return mv.Interface(), nil
		}
	}

	return nil, fmt.Errorf("invalid token reference %q", token)
}

func (r refResolver) loadFrom(ref string, val interface{}) error {
	dir := filepath.Dir(r.path)
	if !filepath.IsAbs(ref) {
		ref = filepath.Join(dir, ref)
	}

	err := r.reader.Read(ref, val, r.eh)
	if err != nil {
		return err
	}

	return nil
}

func isLocalRef(s string) bool {
	return strings.HasPrefix(s, "#")
}

func isSingleElem(s string) bool {
	return strings.Contains(s, "#")
}
