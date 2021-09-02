package asyncApi

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/openapi"
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

type resolver interface {
	Resolve(token string) (interface{}, error)
}

func (r refResolver) resolveConfig() error {
	for _, ch := range r.config.Channels {
		if err := r.resolveChannelRef(ch); err != nil {
			return err
		}
	}

	return nil
}

func (r refResolver) resolveChannelRef(m *ChannelRef) error {
	if m == nil {
		return nil
	}

	if len(m.Ref) > 0 && m.Value == nil {
		if err := r.resolve(m.Ref, r.config, &m.Value); err != nil {
			return err
		}
	}

	if m.Value == nil {
		return nil
	}

	if err := r.resolveMessageRef(m.Value.Publish.Message); err != nil {
		return err
	}
	if err := r.resolveMessageRef(m.Value.Subscribe.Message); err != nil {
		return err
	}

	return nil
}

func (r refResolver) resolveMessageRef(m *MessageRef) error {
	if m == nil {
		return nil
	}

	if len(m.Ref) > 0 && m.Value == nil {
		if err := r.resolve(m.Ref, r.config, &m.Value); err != nil {
			return err
		}
	}

	if m.Value == nil {
		return nil
	}

	return r.resolveSchemaRef(m.Value.Payload)
}

func (r refResolver) resolveSchemas(s *openapi.Schemas) error {
	if s == nil {
		return nil
	}

	if len(s.Ref) > 0 && s.Value == nil {
		if err := r.resolve(s.Ref, r.config, &s.Value); err != nil {
			return err
		}
	}

	if s.Value == nil {
		return nil
	}

	for _, child := range s.Value {
		if err := r.resolveSchemaRef(child); err != nil {
			return err
		}
	}

	return nil
}

func (r refResolver) resolveSchemaRef(s *openapi.SchemaRef) error {
	if s == nil {
		return nil
	}

	if len(s.Ref) > 0 && s.Value == nil {
		resolved := &openapi.SchemaRef{}
		if err := r.resolve(s.Ref, r.config, &resolved); err != nil {
			return err
		}
		s.Value = resolved.Value
	}

	if s.Value == nil {
		return nil
	}

	if err := r.resolveSchemaRef(s.Value.Items); err != nil {
		return err
	}

	if err := r.resolveSchemas(s.Value.Properties); err != nil {
		return err
	}

	if err := r.resolveSchemaRef(s.Value.AdditionalProperties); err != nil {
		return err
	}

	return nil
}

func get(token string, node interface{}) (interface{}, error) {
	rValue := reflect.Indirect(reflect.ValueOf(node))

	if r, ok := node.(resolver); ok {
		return r.Resolve(token)
	}

	switch rValue.Kind() {
	case reflect.Struct:
		f := caseInsenstiveFieldByName(rValue, token)
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

func (r *MessageRef) resolve(token string) (interface{}, error) {
	return get(token, r.Value)
}

func caseInsenstiveFieldByName(v reflect.Value, name string) reflect.Value {
	name = strings.ToLower(name)
	return v.FieldByNameFunc(func(n string) bool { return strings.ToLower(n) == name })
}

func (r refResolver) resolve(ref string, config interface{}, val interface{}) (err error) {
	u, err := url.Parse(ref)
	if err != nil {
		return err
	}

	if len(u.Path) > 0 {
		switch s := strings.ToLower(u.Fragment); {
		case strings.HasPrefix(s, "/components"):
			var c *Config
			err = r.readConfig(u.Path, &c)
			config = c
		case len(s) == 0:
			err = r.readConfig(u.Path, val)
			config = val
		default:
			switch val.(type) {
			case **openapi.SchemaRef:
				schemas := &openapi.Schemas{}
				err = r.readConfig(u.Path, &schemas.Value)
				config = schemas
			}

		}

		if err != nil {
			return err
		}
	}

	tokens := strings.Split(u.Fragment, "/")

	for _, t := range tokens[1:] {
		config, err = get(t, config)
		if err != nil {
			return
		}
	}

	if config == nil {
		return fmt.Errorf("found unresolved ref: %q", ref)
	}

	if reflect.ValueOf(config).Kind() == reflect.Ptr {
		i := 2
		_ = i
	}

	v := reflect.ValueOf(config)
	if reflect.Indirect(v).Kind() == reflect.Map {
		reflect.Indirect(reflect.ValueOf(val)).Set(reflect.Indirect(v))
		return
	}
	v2 := reflect.Indirect(reflect.ValueOf(val))
	if !v.Type().AssignableTo(v2.Type()) {
		v = v.Elem()
	}
	v2.Set(v)

	return
}

func (r refResolver) readConfig(path string, node interface{}) error {
	dir := filepath.Dir(r.path)
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}

	err := r.reader.Read(path, node, r.eh)
	return err
}
