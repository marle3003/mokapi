package asyncApi

import (
	"mokapi/config/dynamic"
)

func (c *Config) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	for _, ch := range c.Channels {
		if ch == nil {
			continue
		}
		if err := ch.Parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (c *ChannelRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(c.Ref) > 0 {
		return dynamic.Resolve(c.Ref, &c.Value, config, reader)
	}

	if c.Value == nil {
		return nil
	}

	if c.Value.Publish != nil {
		if err := c.Value.Publish.Parse(config, reader); err != nil {
			return err
		}
	}

	if c.Value.Subscribe != nil {
		if err := c.Value.Subscribe.Parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (o *Operation) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if o.Message != nil {
		return o.Message.Parse(config, reader)
	}
	return nil
}

func (r *MessageRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}

	if r.Value == nil {
		return nil
	}

	if r.Value.Payload != nil {
		if err := r.Value.Payload.Parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}
