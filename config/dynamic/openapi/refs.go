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
	Resolve(token string) (interface{}, error)
}

func (r refResolver) resolveConfig() error {
	if err := r.resolveSchemas(r.config.Components.Schemas); err != nil {
		return err
	}

	if err := r.resolveResponses(r.config.Components.Responses); err != nil {
		return err
	}

	if err := r.resolveRequestBodies(r.config.Components.RequestBodies); err != nil {
		return err
	}

	if err := r.resolveParameters(r.config.Components.Parameters); err != nil {
		return err
	}

	if err := r.resolveExamples(r.config.Components.Examples); err != nil {
		return err
	}

	if err := r.resolveHeaders(r.config.Components.Headers); err != nil {
		return err
	}

	for _, e := range r.config.EndPoints {
		if err := r.resolveEndpointRef(e); err != nil {
			return err
		}
	}

	return nil
}

func (r refResolver) resolveSchemas(s *Schemas) error {
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

func (r refResolver) resolveResponses(res *NamedResponses) error {
	if res == nil {
		return nil
	}

	if len(res.Ref) > 0 && res.Value == nil {
		if err := r.resolve(res.Ref, r.config, &res.Value); err != nil {
			return err
		}
	}

	return nil
}

func (r refResolver) resolveRequestBodies(req *RequestBodies) error {
	if req == nil {
		return nil
	}

	if len(req.Ref) > 0 && req.Value == nil {
		if err := r.resolve(req.Ref, r.config, &req.Value); err != nil {
			return err
		}
	}

	return nil
}

func (r refResolver) resolveParameters(p *NamedParameters) error {
	if p == nil {
		return nil
	}

	if len(p.Ref) > 0 && p.Value == nil {
		if err := r.resolve(p.Ref, r.config, &p.Value); err != nil {
			return err
		}
	}

	return nil
}

func (r refResolver) resolveExamples(e *Examples) error {
	if e == nil {
		return nil
	}

	if len(e.Ref) > 0 && e.Value == nil {
		if err := r.resolve(e.Ref, r.config, &e.Value); err != nil {
			return err
		}
	}

	return nil
}

func (r refResolver) resolveHeaders(e *NamedHeaders) error {
	if e == nil {
		return nil
	}

	if len(e.Ref) > 0 && e.Value == nil {
		if err := r.resolve(e.Ref, r.config, &e.Value); err != nil {
			return err
		}
	}

	return nil
}

func (r refResolver) resolveEndpointRef(e *EndpointRef) error {
	if e == nil {
		return nil
	}

	if len(e.Ref) > 0 && e.Value == nil {
		//var resolved *EndpointRef
		if err := r.resolve(e.Ref, r.config, &e.Value); err != nil {
			return err
		}
		//e.Value = resolved.Value
	}

	for _, p := range e.Value.Parameters {
		if err := r.resolveParameter(p); err != nil {
			return err
		}
	}

	for _, o := range e.Value.Operations() {
		o.Endpoint = e.Value
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
		var resolved *ParameterRef
		if err := r.resolve(p.Ref, r.config, &resolved); err != nil {
			return err
		}
		p.Value = resolved.Value
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
		var resolved *RequestBodyRef
		if err := r.resolve(req.Ref, r.config, &resolved); err != nil {
			return err
		}
		req.Value = resolved.Value
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
		var resolved *ResponseRef
		if err := r.resolve(res.Ref, r.config, &resolved); err != nil {
			return err
		}
		res.Value = resolved.Value
	}

	if res.Value == nil {
		return nil
	}

	for _, h := range res.Value.Headers {
		if err := r.resolveHeader(h); err != nil {
			return err
		}
	}

	for _, c := range res.Value.Content {
		if c == nil {
			continue
		}
		if err := r.resolveSchemaRef(c.Schema); err != nil {
			return err
		}

		for _, e := range c.Examples {
			if err := r.resolveExample(e); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r refResolver) resolveSchemaRef(s *SchemaRef) error {
	if s == nil {
		return nil
	}

	if len(s.Ref) > 0 && s.Value == nil {
		//var resolved *SchemaRef
		resolved := &SchemaRef{}
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

func (r refResolver) resolveExample(ref *ExampleRef) error {
	if ref == nil {
		return nil
	}

	if len(ref.Ref) > 0 && ref.Value == nil {
		var resolved *ExampleRef
		if err := r.resolve(ref.Ref, r.config, &resolved); err != nil {
			return err
		}
		ref.Value = resolved.Value
	}

	return nil
}

func (r refResolver) resolveHeader(ref *HeaderRef) error {
	if ref == nil {
		return nil
	}

	if len(ref.Ref) > 0 && ref.Value == nil {
		var resolved *HeaderRef
		if err := r.resolve(ref.Ref, r.config, &resolved); err != nil {
			return err
		}
		ref.Value = resolved.Value
	}

	return nil
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
			case **SchemaRef:
				schemas := &Schemas{}
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

func caseInsenstiveFieldByName(v reflect.Value, name string) reflect.Value {
	name = strings.ToLower(name)
	return v.FieldByNameFunc(func(n string) bool { return strings.ToLower(n) == name })
}

func (r refResolver) readConfig(path string, node interface{}) error {
	dir := filepath.Dir(r.path)
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}

	err := r.reader.Read(path, node, r.eh)
	return err
}

func (s *Schemas) Resolve(token string) (interface{}, error) {
	return get(token, s.Value)
}

func (r *NamedResponses) Resolve(token string) (interface{}, error) {
	return get(token, r.Value)
}

func (r *NamedParameters) Resolve(token string) (interface{}, error) {
	return get(token, r.Value)
}

func (r *Examples) Resolve(token string) (interface{}, error) {
	return get(token, r.Value)
}

func (r *RequestBodies) Resolve(token string) (interface{}, error) {
	return get(token, r.Value)
}

func (r *NamedHeaders) Resolve(token string) (interface{}, error) {
	return get(token, r.Value)
}
