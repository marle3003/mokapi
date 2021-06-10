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

	return r.resolveSchemaRef(m.Value.Payload)
}

func (r refResolver) resolveSchemas(s *openapi.Schemas) error {
	if s == nil {
		return nil
	}

	if len(s.Ref) > 0 && s.Value == nil {
		u, err := url.Parse(s.Ref)
		if err != nil {
			return err
		}

		if !isLocalRef(s.Ref) {
			err := r.loadFrom(u.Path, &s.Value)
			if err != nil {
				return err
			}
		} else {
			err := r.resolve(u.Fragment, r.config, &s.Value)
			if err != nil {
				return err
			}
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
		u, err := url.Parse(s.Ref)
		if err != nil {
			return err
		}

		if !isLocalRef(s.Ref) {
			if len(u.Fragment) == 0 {
				err := r.loadFrom(u.Path, &s.Value)
				if err != nil {
					return err
				}
			} else {
				schemas := &openapi.Schemas{}
				err := r.loadFrom(u.Path, &schemas.Value)
				if err != nil {
					return err
				}
				var resolved *openapi.SchemaRef
				if err := r.resolve(u.Fragment, schemas.Value, &resolved); err != nil {
					return err
				}
				s.Value = resolved.Value
			}

		} else {
			schema := &openapi.SchemaRef{}
			err := r.resolve(u.Fragment, r.config, &schema)
			if err != nil {
				return err
			}
			s.Value = schema.Value
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

func (r refResolver) resolve(ref string, node interface{}, val interface{}) (err error) {
	tokens := strings.Split(ref, "/")

	i := node
	for _, t := range tokens[1:] {
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

func (r *MessageRef) resolve(token string) (interface{}, error) {
	return get(token, r.Value)
}

func caseInsenstiveFieldByName(v reflect.Value, name string) reflect.Value {
	name = strings.ToLower(name)
	return v.FieldByNameFunc(func(n string) bool { return strings.ToLower(n) == name })
}
