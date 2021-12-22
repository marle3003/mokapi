package openapi

import (
	"mokapi/config/dynamic/common"
)

func (c *Config) Parse(file *common.File, reader common.Reader) error {
	if c == nil {
		return nil
	}

	if err := c.Components.Schemas.Parse(file, reader); err != nil {
		return err
	}

	if err := c.Components.Responses.Parse(file, reader); err != nil {
		return err
	}

	if err := c.Components.RequestBodies.Parse(file, reader); err != nil {
		return err
	}

	if err := c.Components.Parameters.Parse(file, reader); err != nil {
		return err
	}

	if err := c.Components.Examples.Parse(file, reader); err != nil {
		return err
	}

	if err := c.Components.Headers.Parse(file, reader); err != nil {
		return err
	}

	for _, e := range c.EndPoints {
		if err := e.Parse(file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (s *Schemas) Parse(file *common.File, reader common.Reader) error {
	if s == nil {
		return nil
	}
	if len(s.Ref) > 0 && s.Value == nil {
		if err := common.Resolve(s.Ref, &s.Value, file, reader); err != nil {
			return err
		}
	}

	if s.Value == nil {
		return nil
	}

	for _, child := range s.Value {
		if err := child.Parse(file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (s *SchemaRef) Parse(file *common.File, reader common.Reader) error {
	if s == nil {
		return nil
	}
	if len(s.Ref) > 0 && s.Value == nil {
		if err := common.Resolve(s.Ref, &s.Value, file, reader); err != nil {
			return err
		}
	}

	if s.Value == nil {
		return nil
	}

	return s.Value.Parse(file, reader)
}

func (s *Schema) Parse(file *common.File, reader common.Reader) error {
	if s == nil {
		return nil
	}

	if err := s.Items.Parse(file, reader); err != nil {
		return err
	}

	if err := s.Properties.Parse(file, reader); err != nil {
		return err
	}

	if err := s.AdditionalProperties.Parse(file, reader); err != nil {
		return err
	}

	return nil
}

func (res *NamedResponses) Parse(file *common.File, reader common.Reader) error {
	if res == nil {
		return nil
	}

	if len(res.Ref) > 0 && res.Value == nil {
		if err := common.Resolve(res.Ref, &res.Value, file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (req *RequestBodies) Parse(file *common.File, reader common.Reader) error {
	if req == nil {
		return nil
	}

	if len(req.Ref) > 0 && req.Value == nil {
		if err := common.Resolve(req.Ref, &req.Value, file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (p *NamedParameters) Parse(file *common.File, reader common.Reader) error {
	if p == nil {
		return nil
	}

	if len(p.Ref) > 0 && p.Value == nil {
		if err := common.Resolve(p.Ref, &p.Value, file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (e *Examples) Parse(file *common.File, reader common.Reader) error {
	if e == nil {
		return nil
	}

	if len(e.Ref) > 0 && e.Value == nil {
		if err := common.Resolve(e.Ref, &e.Value, file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (h *NamedHeaders) Parse(file *common.File, reader common.Reader) error {
	if h == nil {
		return nil
	}

	if len(h.Ref) > 0 && h.Value == nil {
		if err := common.Resolve(h.Ref, &h.Value, file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (e *EndpointRef) Parse(file *common.File, reader common.Reader) error {
	if e == nil {
		return nil
	}

	if len(e.Ref) > 0 && e.Value == nil {
		if err := common.Resolve(e.Ref, &e.Value, file, reader); err != nil {
			return err
		}
	}

	return e.Value.Parse(file, reader)
}

func (e *Endpoint) Parse(file *common.File, reader common.Reader) error {
	if e == nil {
		return nil
	}

	for _, p := range e.Parameters {
		if err := p.Parse(file, reader); err != nil {
			return err
		}
	}

	for _, o := range e.Operations() {
		o.Endpoint = e
		for _, p := range o.Parameters {
			if err := p.Parse(file, reader); err != nil {
				return err
			}
		}

		if err := o.RequestBody.Parse(file, reader); err != nil {
			return err
		}

		for it := o.Responses.Iter(); it.Next(); {
			res := it.Value().(*ResponseRef)
			if err := res.Parse(file, reader); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *ParameterRef) Parse(file *common.File, reader common.Reader) error {
	if p == nil {
		return nil
	}

	if len(p.Ref) > 0 && p.Value == nil {
		if err := common.Resolve(p.Ref, &p.Value, file, reader); err != nil {
			return err
		}
	}

	if p.Value == nil {
		return nil
	}

	if err := p.Value.Schema.Parse(file, reader); err != nil {
		return err
	}

	return nil
}

func (req *RequestBodyRef) Parse(file *common.File, reader common.Reader) error {
	if req == nil {
		return nil
	}

	if len(req.Ref) > 0 && req.Value == nil {
		if err := common.Resolve(req.Ref, &req.Value, file, reader); err != nil {
			return err
		}
	}

	for _, c := range req.Value.Content {
		if c == nil {
			continue
		}
		if err := c.Schema.Parse(file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (res *ResponseRef) Parse(file *common.File, reader common.Reader) error {
	if res == nil {
		return nil
	}

	if len(res.Ref) > 0 && res.Value == nil {
		if err := common.Resolve(res.Ref, &res.Value, file, reader); err != nil {
			return err
		}
	}

	if res.Value == nil {
		return nil
	}

	for _, h := range res.Value.Headers {
		if err := h.Parse(file, reader); err != nil {
			return err
		}
	}

	for _, c := range res.Value.Content {
		if c == nil {
			continue
		}
		if err := c.Schema.Parse(file, reader); err != nil {
			return err
		}

		for _, e := range c.Examples {
			if err := e.Parse(file, reader); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *ExampleRef) Parse(file *common.File, reader common.Reader) error {
	if e == nil {
		return nil
	}

	if len(e.Ref) > 0 && e.Value == nil {
		if err := common.Resolve(e.Ref, &e.Value, file, reader); err != nil {
			return err
		}
	}
	return nil
}

func (h *HeaderRef) Parse(file *common.File, reader common.Reader) error {
	if h == nil {
		return nil
	}

	if len(h.Ref) > 0 && h.Value == nil {
		if err := common.Resolve(h.Ref, &h.Value, file, reader); err != nil {
			return err
		}
	}
	return nil
}

// ------
//
//type ReferenceResolver struct {
//	reader dynamic.Reader
//	file   *dynamic.File
//	config *Config
//}
//
//func resolve(file *dynamic.File, reader dynamic.Reader) error {
//	config, ok := file.Data.(*Config)
//	if !ok {
//		return fmt.Errorf("unexpected config %v", reflect.TypeOf(file.Data).String())
//	}
//
//	r := &ReferenceResolver{
//		reader: reader,
//		config: config,
//		file:   file,
//	}
//
//	return r.Resolve()
//}
//
//func (r ReferenceResolver) Resolve() error {
//	if err := r.resolveSchemas(r.config.Components.Schemas); err != nil {
//		return err
//	}
//
//	if err := r.resolveResponses(r.config.Components.Responses); err != nil {
//		return err
//	}
//
//	if err := r.resolveRequestBodies(r.config.Components.RequestBodies); err != nil {
//		return err
//	}
//
//	if err := r.resolveParameters(r.config.Components.Parameters); err != nil {
//		return err
//	}
//
//	if err := r.resolveExamples(r.config.Components.Examples); err != nil {
//		return err
//	}
//
//	if err := r.resolveHeaders(r.config.Components.Headers); err != nil {
//		return err
//	}
//
//	for _, e := range r.config.EndPoints {
//		if err := r.resolveEndpointRef(e); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveSchemas(s *Schemas) error {
//	if s == nil {
//		return nil
//	}
//
//	if len(s.Ref) > 0 && s.Value == nil {
//		if err := r.resolve(s.Ref, &s.Value); err != nil {
//			return err
//		}
//	}
//
//	if s.Value == nil {
//		return nil
//	}
//
//	for _, child := range s.Value {
//		if err := r.resolveSchemaRef(child); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveResponses(res *NamedResponses) error {
//	if res == nil {
//		return nil
//	}
//
//	if len(res.Ref) > 0 && res.Value == nil {
//		if err := r.resolve(res.Ref, &res.Value); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveRequestBodies(req *RequestBodies) error {
//	if req == nil {
//		return nil
//	}
//
//	if len(req.Ref) > 0 && req.Value == nil {
//		if err := r.resolve(req.Ref, &req.Value); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveParameters(p *NamedParameters) error {
//	if p == nil {
//		return nil
//	}
//
//	if len(p.Ref) > 0 && p.Value == nil {
//		if err := r.resolve(p.Ref, &p.Value); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveExamples(e *Examples) error {
//	if e == nil {
//		return nil
//	}
//
//	if len(e.Ref) > 0 && e.Value == nil {
//		if err := r.resolve(e.Ref, &e.Value); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveHeaders(e *NamedHeaders) error {
//	if e == nil {
//		return nil
//	}
//
//	if len(e.Ref) > 0 && e.Value == nil {
//		if err := r.resolve(e.Ref, &e.Value); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveEndpointRef(e *EndpointRef) error {
//	if e == nil {
//		return nil
//	}
//
//	if len(e.Ref) > 0 && e.Value == nil {
//		if err := r.resolve(e.Ref, &e.Value); err != nil {
//			return err
//		}
//	}
//
//	for _, p := range e.Value.Parameters {
//		if err := r.resolveParameter(p); err != nil {
//			return err
//		}
//	}
//
//	for _, o := range e.Value.Operations() {
//		o.Endpoint = e.Value
//		for _, p := range o.Parameters {
//			if err := r.resolveParameter(p); err != nil {
//				return err
//			}
//		}
//
//		if err := r.resolveRequestBodyRef(o.RequestBody); err != nil {
//			return err
//		}
//
//		for _, res := range o.Responses {
//			if err := r.resolveResponseRef(res); err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveParameter(p *ParameterRef) error {
//	if p == nil {
//		return nil
//	}
//
//	if len(p.Ref) > 0 && p.Value == nil {
//		if err := r.resolve(p.Ref, &p.Value); err != nil {
//			return err
//		}
//	}
//
//	if p.Value == nil {
//		return nil
//	}
//
//	if err := r.resolveSchemaRef(p.Value.Schema); err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveRequestBodyRef(req *RequestBodyRef) error {
//	if req == nil {
//		return nil
//	}
//
//	if len(req.Ref) > 0 && req.Value == nil {
//		if err := r.resolve(req.Ref, &req.Value); err != nil {
//			return err
//		}
//	}
//
//	for _, c := range req.Value.Content {
//		if c == nil {
//			continue
//		}
//		if err := r.resolveSchemaRef(c.Schema); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveResponseRef(res *ResponseRef) error {
//	if res == nil {
//		return nil
//	}
//
//	if len(res.Ref) > 0 && res.Value == nil {
//		if err := r.resolve(res.Ref, &res.Value); err != nil {
//			return err
//		}
//	}
//
//	if res.Value == nil {
//		return nil
//	}
//
//	for _, h := range res.Value.Headers {
//		if err := r.resolveHeader(h); err != nil {
//			return err
//		}
//	}
//
//	for _, c := range res.Value.Content {
//		if c == nil {
//			continue
//		}
//		if err := r.resolveSchemaRef(c.Schema); err != nil {
//			return err
//		}
//
//		for _, e := range c.Examples {
//			if err := r.resolveExample(e); err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveSchemaRef(s *SchemaRef) error {
//	if s == nil {
//		return nil
//	}
//
//	if len(s.Ref) > 0 && s.Value == nil {
//		if err := r.resolve(s.Ref, &s.Value); err != nil {
//			return err
//		}
//	}
//
//	if s.Value == nil {
//		return nil
//	}
//
//	if err := r.resolveSchemaRef(s.Value.Items); err != nil {
//		return err
//	}
//
//	if err := r.resolveSchemas(s.Value.Properties); err != nil {
//		return err
//	}
//
//	if err := r.resolveSchemaRef(s.Value.AdditionalProperties); err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveExample(ref *ExampleRef) error {
//	if ref == nil {
//		return nil
//	}
//
//	if len(ref.Ref) > 0 && ref.Value == nil {
//		if err := r.resolve(ref.Ref, &ref.Value); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveHeader(ref *HeaderRef) error {
//	if ref == nil {
//		return nil
//	}
//
//	if len(ref.Ref) > 0 && ref.Value == nil {
//		if err := r.resolve(ref.Ref, &ref.Value); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolve(ref string, val interface{}) (err error) {
//	u, err := url.Parse(ref)
//	if err != nil {
//		return err
//	}
//
//	var i interface{}
//	if len(u.Path) > 0 {
//		if !u.IsAbs() {
//			u, err = r.file.Url.Parse(ref)
//		}
//
//		// TODO each object should implement resolve
//		// ReferenceResolver should be a parameter of that resolve func
//		// switch below should not be needed anymore, just call file.Data.resolve
//		// if file.data implements a specific interface
//
//		var f *dynamic.File
//		switch s := strings.ToLower(u.Fragment); {
//		case strings.HasPrefix(s, "/components"):
//			f, err = r.reader.Read(u, dynamic.WithData(&Config{}), dynamic.WithParent(r.file), dynamic.WithInitializer(func(file *dynamic.File, reader dynamic.Reader) error {
//				err := resolve(file, reader)
//				if err != nil {
//					return err
//				}
//				return common.ResolvePath(u.Fragment, file.Data, val)
//			}))
//			return err
//		case len(s) == 0:
//			f, err = r.reader.Read(u, dynamic.WithData(val), dynamic.WithParent(r.file), dynamic.WithInitializer(func(file *dynamic.File, reader dynamic.Reader) error {
//				// TODO
//				//err := resolve(file, reader)
//				//if err != nil {
//				//	return err
//				//}
//				return common.ResolvePath(u.Fragment, file.Data, val)
//			}))
//		default:
//			switch val.(type) {
//			case **SchemaRef:
//				schemas := &Schemas{}
//				f, err = r.reader.Read(u, dynamic.WithData(schemas.Value))
//			}
//
//		}
//
//		if err != nil {
//			return
//		}
//		i = f.Data
//	} else {
//		i = r.config
//	}
//
//	return common.ResolvePath(u.Fragment, i, val)
//}
//
//func (s *Schemas) Resolve(token string) (interface{}, error) {
//	return common.Get(token, s.Value)
//}
//
//func (r *NamedResponses) Resolve(token string) (interface{}, error) {
//	return common.Get(token, r.Value)
//}
//
//func (r *NamedParameters) Resolve(token string) (interface{}, error) {
//	return common.Get(token, r.Value)
//}
//
//func (r *Examples) Resolve(token string) (interface{}, error) {
//	return common.Get(token, r.Value)
//}
//
//func (r *RequestBodies) Resolve(token string) (interface{}, error) {
//	return common.Get(token, r.Value)
//}
//
//func (r *NamedHeaders) Resolve(token string) (interface{}, error) {
//	return common.Get(token, r.Value)
//}
//
//func (r *SchemaRef) Resolve(token string) (interface{}, error) {
//	return common.Get(token, r.Value)
//}
