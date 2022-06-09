package asyncApi

import (
	"mokapi/config/dynamic/common"
)

func (c *Config) Parse(config *common.Config, reader common.Reader) error {
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

func (c *ChannelRef) Parse(config *common.Config, reader common.Reader) error {
	if len(c.Ref) > 0 {
		return common.Resolve(c.Ref, &c.Value, config, reader)
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

func (o *Operation) Parse(config *common.Config, reader common.Reader) error {
	if o.Message != nil {
		return o.Message.Parse(config, reader)
	}
	return nil
}

func (r *MessageRef) Parse(config *common.Config, reader common.Reader) error {
	if len(r.Ref) > 0 {
		if err := common.Resolve(r.Ref, &r.Value, config, reader); err != nil {
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
