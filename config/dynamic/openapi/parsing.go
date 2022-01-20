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

func (r *NamedResponses) Parse(file *common.File, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref()) > 0 && r.Value == nil {
		if err := common.Resolve(r.Ref(), &r.Value, file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *RequestBodies) Parse(file *common.File, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref()) > 0 && r.Value == nil {
		if err := common.Resolve(r.Ref(), &r.Value, file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *Examples) Parse(file *common.File, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref()) > 0 && r.Value == nil {
		if err := common.Resolve(r.Ref(), &r.Value, file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *NamedHeaders) Parse(file *common.File, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref()) > 0 && r.Value == nil {
		if err := common.Resolve(r.Ref(), &r.Value, file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (e *EndpointRef) Parse(file *common.File, reader common.Reader) error {
	if e == nil {
		return nil
	}

	if len(e.Ref()) > 0 && e.Value == nil {
		if err := common.Resolve(e.Ref(), &e.Value, file, reader); err != nil {
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

		if o.Responses != nil {
			for it := o.Responses.Iter(); it.Next(); {
				res := it.Value().(*ResponseRef)
				if err := res.Parse(file, reader); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *RequestBodyRef) Parse(file *common.File, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref()) > 0 && r.Value == nil {
		if err := common.Resolve(r.Ref(), &r.Value, file, reader); err != nil {
			return err
		}
	}

	for _, c := range r.Value.Content {
		if c == nil {
			continue
		}
		if err := c.Schema.Parse(file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *ResponseRef) Parse(file *common.File, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref()) > 0 && r.Value == nil {
		if err := common.Resolve(r.Ref(), &r.Value, file, reader); err != nil {
			return err
		}
	}

	if r.Value == nil {
		return nil
	}

	for _, h := range r.Value.Headers {
		if err := h.Parse(file, reader); err != nil {
			return err
		}
	}

	for _, c := range r.Value.Content {
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

func (r *ExampleRef) Parse(file *common.File, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref()) > 0 && r.Value == nil {
		if err := common.Resolve(r.Ref(), &r.Value, file, reader); err != nil {
			return err
		}
	}
	return nil
}

func (r *HeaderRef) Parse(file *common.File, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref()) > 0 && r.Value == nil {
		if err := common.Resolve(r.Ref(), &r.Value, file, reader); err != nil {
			return err
		}
	}
	return nil
}
