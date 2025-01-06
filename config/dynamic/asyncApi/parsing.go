package asyncApi

import (
	"mokapi/config/dynamic"
)

func (c *Config) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	for _, server := range c.Servers {
		if server == nil || len(server.Ref) == 0 {
			continue
		}
		if err := dynamic.Resolve(server.Ref, &server.Value, config, reader); err != nil {
			return err
		}
	}

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

	for _, param := range c.Value.Parameters {
		if err := param.Parse(config, reader); err != nil {
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

	for _, trait := range r.Value.Traits {
		if err := trait.parse(config, reader); err != nil {
			return err
		}
		r.Value.applyTrait(trait.Value)
	}

	return nil
}

func (r *MessageTraitRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}

	if r.Value == nil {
		return nil
	}

	if r.Value.Headers != nil {
		if err := r.Value.Headers.Parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (m *Message) applyTrait(trait *MessageTrait) {
	if len(trait.MessageId) > 0 {
		m.MessageId = trait.MessageId
	}
	if len(trait.Name) > 0 {
		m.Name = trait.Name
	}
	if len(trait.Title) > 0 {
		m.Title = trait.Title
	}
	if len(trait.Summary) > 0 {
		m.Summary = trait.Summary
	}
	if len(trait.Description) > 0 {
		m.Description = trait.Description
	}
	if len(trait.ContentType) > 0 {
		m.ContentType = trait.ContentType
	}
	if trait.Headers != nil {
		m.Headers = trait.Headers
	}
	if trait.Bindings.Kafka.Key != nil {
		m.Bindings.Kafka.Key = trait.Bindings.Kafka.Key
	}
}

func (r *ParameterRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}
	return nil
}
