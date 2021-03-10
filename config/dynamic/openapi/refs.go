package openapi

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

type resolver interface {
	resolve(token string) (interface{}, error)
}

func (r refResolver) resolveConfig() error {

	if err := r.resolveMokapiRef(r.config.Info.Mokapi); err != nil {
		return err
	}

	if r.config.Components.Schemas != nil {
		if err := r.resolveSchemas(r.config.Components.Schemas); err != nil {
			return err
		}
	}

	if r.config.Components.Responses != nil {
		if err := r.resolveResponses(r.config.Components.Responses); err != nil {
			return err
		}
	}

	if r.config.Components.RequestBodies != nil {
		if err := r.resolveRequestBodies(r.config.Components.RequestBodies); err != nil {
			return err
		}
	}

	for _, e := range r.config.EndPoints {
		if err := r.resolveEndpointRef(e); err != nil {
			return err
		}
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

func (r refResolver) resolveSchemas(s *Schemas) error {
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

	return nil
}

func (r refResolver) resolveResponses(res *NamedResponses) error {
	if res == nil {
		return nil
	}

	if len(res.Ref) > 0 && res.Value == nil {
		u, err := url.Parse(res.Ref)
		if err != nil {
			return err
		}

		if !isLocalRef(res.Ref) {
			err := r.loadFrom(u.Path, &res.Value)
			if err != nil {
				return err
			}
		} else {
			err := r.resolve(u.Fragment, r.config, &res.Value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r refResolver) resolveRequestBodies(req *RequestBodies) error {
	if req == nil {
		return nil
	}

	if len(req.Ref) > 0 && req.Value == nil {
		u, err := url.Parse(req.Ref)
		if err != nil {
			return err
		}

		if !isLocalRef(req.Ref) {
			err := r.loadFrom(u.Path, &req.Value)
			if err != nil {
				return err
			}
		} else {
			err := r.resolve(u.Fragment, r.config, &req.Value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r refResolver) resolveEndpointRef(e *EndpointRef) error {
	if e == nil {
		return nil
	}

	if len(e.Ref) > 0 && e.Value == nil {
		u, err := url.Parse(e.Ref)
		if err != nil {
			return err
		}

		if !isLocalRef(e.Ref) {
			err := r.loadFrom(u.Path, &e.Value)
			if err != nil {
				return err
			}
		} else {
			err := r.resolve(u.Fragment, r.config, &e.Value)
			if err != nil {
				return err
			}
		}
	}

	for _, p := range e.Value.Parameters {
		if err := r.resolveParameter(p); err != nil {
			return err
		}
	}

	for _, o := range e.Value.Operations() {

		for _, p := range o.Parameters {
			if err := r.resolveParameter(p); err != nil {
				return err
			}
		}

		if err := r.resolveRequestBodyRef(o.RequestBody); err != nil {
			return err
		}

		for _, res := range o.Responses {
			if err := r.resolveResponseRef(res); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r refResolver) resolveParameter(p *ParameterRef) error {
	if p == nil {
		return nil
	}

	if len(p.Ref) > 0 && p.Value == nil {
		u, err := url.Parse(p.Ref)
		if err != nil {
			return err
		}

		if !isLocalRef(p.Ref) {
			err := r.loadFrom(u.Path, &p.Value)
			if err != nil {
				return err
			}
		} else {
			err := r.resolve(u.Fragment, r.config, &p.Value)
			if err != nil {
				return err
			}
		}
	}

	if p.Value == nil {
		return nil
	}

	if err := r.resolveSchemaRef(p.Value.Schema); err != nil {
		return err
	}

	return nil
}

func (r refResolver) resolveRequestBodyRef(req *RequestBodyRef) error {
	if req == nil {
		return nil
	}

	if len(req.Ref) > 0 && req.Value == nil {
		u, err := url.Parse(req.Ref)
		if err != nil {
			return err
		}

		if !isLocalRef(req.Ref) {
			err := r.loadFrom(u.Path, &req.Value)
			if err != nil {
				return err
			}
		} else {
			err := r.resolve(u.Fragment, r.config, &req.Value)
			if err != nil {
				return err
			}
		}
	}

	for _, c := range req.Value.Content {
		if c == nil {
			continue
		}
		if err := r.resolveSchemaRef(c.Schema); err != nil {
			return err
		}
	}

	return nil
}

func (r refResolver) resolveResponseRef(res *ResponseRef) error {
	if res == nil {
		return nil
	}

	if len(res.Ref) > 0 && res.Value == nil {
		u, err := url.Parse(res.Ref)
		if err != nil {
			return err
		}

		if !isLocalRef(res.Ref) {
			err := r.loadFrom(u.Path, &res.Value)
			if err != nil {
				return err
			}
		} else {
			err := r.resolve(u.Fragment, r.config, &res.Value)
			if err != nil {
				return err
			}
		}
	}

	for _, c := range res.Value.Content {
		if c == nil {
			continue
		}
		if err := r.resolveSchemaRef(c.Schema); err != nil {
			return err
		}
	}

	return nil
}

func (r refResolver) resolveSchemaRef(s *SchemaRef) error {
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
				schemas := &Schemas{}
				err := r.loadFrom(u.Path, &schemas.Value)
				if err != nil {
					return err
				}
				var resolved *SchemaRef
				if err := r.resolve(u.Fragment, schemas.Value, &resolved); err != nil {
					return err
				}
				s.Value = resolved.Value
			}

		} else {
			err := r.resolve(s.Ref, r.config, &s)
			if err != nil {
				return err
			}
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

	if r, ok := node.(resolver); ok {
		return r.resolve(token)
	}

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
