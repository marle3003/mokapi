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

	if err := c.Components.Examples.parse(config, reader); err != nil {
		return err
	}

	if err := c.Components.Headers.parse(config, reader); err != nil {
		return err
	}

	if err := c.Paths.parse(config, reader); err != nil {
		return err
	}

	return nil
}

func (r *Responses) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	return nil
}

func (r *RequestBodies) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
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
