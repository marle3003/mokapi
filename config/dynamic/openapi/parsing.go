package openapi

import (
	"mokapi/config/dynamic/common"
)

func (c *Config) Parse(config *common.Config, reader common.Reader) error {
	if c == nil {
		return nil
	}

	if err := c.Components.Schemas.Parse(config, reader); err != nil {
		return err
	}

	if err := c.Components.Responses.Parse(config, reader); err != nil {
		return err
	}

	if err := c.Components.RequestBodies.Parse(config, reader); err != nil {
		return err
	}

	if err := c.Components.Parameters.Parse(config, reader); err != nil {
		return err
	}

	if err := c.Components.Examples.Parse(config, reader); err != nil {
		return err
	}

	if err := c.Components.Headers.Parse(config, reader); err != nil {
		return err
	}

	if err := c.Paths.parse(config, reader); err != nil {
		return err
	}

	return nil
}

func (r *NamedResponses) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return common.Resolve(r.Ref, &r.Value, config, reader)
	}

	return nil
}

func (r *RequestBodies) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return common.Resolve(r.Ref, &r.Value, config, reader)
	}

	return nil
}

func (r *Examples) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return common.Resolve(r.Ref, &r.Value, config, reader)
	}

	return nil
}

func (r *NamedHeaders) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return common.Resolve(r.Ref, &r.Value, config, reader)
	}

	return nil
}

func (r *RequestBodyRef) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return common.Resolve(r.Ref, &r.Value, config, reader)
	}

	for _, c := range r.Value.Content {
		if c == nil {
			continue
		}
		if err := c.Schema.Parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *ResponseRef) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return common.Resolve(r.Ref, &r.Value, config, reader)
	}

	if r.Value == nil {
		return nil
	}

	for _, h := range r.Value.Headers {
		if err := h.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, c := range r.Value.Content {
		if c == nil {
			continue
		}
		if err := c.Schema.Parse(config, reader); err != nil {
			return err
		}

		for _, e := range c.Examples {
			if err := e.Parse(config, reader); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ExampleRef) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return common.Resolve(r.Ref, &r.Value, config, reader)
	}
	return nil
}

func (r *HeaderRef) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		return common.Resolve(r.Ref, &r.Value, config, reader)
	}
	return nil
}
