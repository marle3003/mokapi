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

type ReferenceResolver struct {
	reader dynamic.ConfigReader
	path   string
	config *Config
	eh     dynamic.ChangeEventHandler
}

func (r ReferenceResolver) ResolveConfig() error {
	for _, ch := range r.config.Channels {
		if err := r.resolveChannelRef(ch); err != nil {
			return err
		}
	}

	return nil
}

func (r ReferenceResolver) resolveChannelRef(m *ChannelRef) error {
	if m == nil {
		return nil
	}

	if len(m.Ref) > 0 && m.Value == nil {
		if resolver, err := r.resolve(m.Ref, r.config, &m.Value); err != nil {
			return err
		} else if !resolver.isEmpty() {
			return resolver.resolveChannelRef(m)
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

func (r ReferenceResolver) resolveMessageRef(m *MessageRef) error {
	if m == nil {
		return nil
	}

	if len(m.Ref) > 0 && m.Value == nil {
		if resolver, err := r.resolve(m.Ref, r.config, &m.Value); err != nil {
			return err
		} else if !resolver.isEmpty() {
			return resolver.resolveMessageRef(m)
		}
	}

	if m.Value == nil {
		return nil
	}

	return r.resolveSchemaRef(m.Value.Payload)
}

func (r ReferenceResolver) resolveSchemas(s *openapi.Schemas) error {
	if s == nil {
		return nil
	}

	if len(s.Ref) > 0 && s.Value == nil {
		if resolver, err := r.resolve(s.Ref, r.config, &s.Value); err != nil {
			return err
		} else if !resolver.isEmpty() {
			return resolver.resolveSchemas(s)
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

func (r ReferenceResolver) resolveSchemaRef(s *openapi.SchemaRef) error {
	if s == nil {
		return nil
	}

	if len(s.Ref) > 0 && s.Value == nil {
		if resolver, err := r.resolve(s.Ref, r.config, &s.Value); err != nil {
			return err
		} else if !resolver.isEmpty() {
			return resolver.resolveSchemaRef(s)
		}
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
	if len(token) == 0 {
		return node, nil
	}

	rValue := reflect.Indirect(reflect.ValueOf(node))

	if r, ok := node.(openapi.Resolver); ok {
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

func caseInsenstiveFieldByName(v reflect.Value, name string) reflect.Value {
	name = strings.ToLower(name)
	return v.FieldByNameFunc(func(n string) bool { return strings.ToLower(n) == name })
}

func (r ReferenceResolver) resolve(ref string, config interface{}, val interface{}) (resolver ReferenceResolver, err error) {
	u, err := url.Parse(ref)
	if err != nil {
		return
	}

	if len(u.Path) > 0 {
		path := u.Path
		if !filepath.IsAbs(u.Path) {
			path = filepath.Join(filepath.Dir(r.path), u.Path)
		}

		resolver = ReferenceResolver{reader: r.reader, path: path, eh: r.eh}

		switch s := strings.ToLower(u.Fragment); {
		case strings.HasPrefix(s, "/components"), strings.HasPrefix(s, "/channels"):
			var c *Config
			err = r.reader.Read(path, &c, r.eh)
			config = c
			resolver.config = c
		case len(s) == 0:
			err = r.reader.Read(path, val, r.eh)
			config = val
		default:
			switch val.(type) {
			case **openapi.SchemaRef:
				schemas := &openapi.Schemas{}
				err = r.reader.Read(path, &schemas.Value, r.eh)
				config = schemas
			}
		}

		if err != nil {
			return
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
		return resolver, fmt.Errorf("found unresolved ref: %q", ref)
	}

	if r, ok := config.(openapi.Resolver); ok {
		if config, err = r.Resolve(""); err != nil {
			return
		}
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

func (r ReferenceResolver) isEmpty() bool {
	return r.reader == nil
}

func (r *MessageRef) Resolve(token string) (interface{}, error) {
	return get(token, r.Value)
}

func (c *ChannelRef) Resolve(token string) (interface{}, error) {
	return get(token, c.Value)
}
