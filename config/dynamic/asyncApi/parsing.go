package asyncApi

import (
	"mokapi/config/dynamic"

	log "github.com/sirupsen/logrus"
)

func (c *Config) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	for _, server := range c.Servers {
		if server == nil || len(server.Ref) == 0 {
			continue
		}
		var resolved *ServerRef
		if err := dynamic.Resolve(server.Ref, &resolved, config, reader); err != nil {
			return err
		}
		server.Value = resolved.Value
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

func (r *ChannelRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		var resolved *ChannelRef
		if err := dynamic.Resolve(r.Ref, &resolved, config, reader); err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
	}

	if r.Value == nil {
		return nil
	}

	if r.Value.Publish != nil {
		if err := r.Value.Publish.Parse(config, reader); err != nil {
			return err
		}
	}

	if r.Value.Subscribe != nil {
		if err := r.Value.Subscribe.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, param := range r.Value.Parameters {
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
		var resolved *MessageRef
		if err := dynamic.Resolve(r.Ref, &resolved, config, reader); err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
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

	if r.Value.ContentType == "" {
		cfg, ok := config.Data.(*Config)
		if ok {
			r.Value.ContentType = cfg.DefaultContentType
		}
		if r.Value.ContentType == "" {
			log.Warn("content type is missing, using application/json")
			r.Value.ContentType = "application/json"
		}
	}

	return nil
}

func (r *MessageTraitRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		var resolved *MessageTraitRef
		if err := dynamic.Resolve(r.Ref, &resolved, config, reader); err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
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
		var resolved *ParameterRef
		if err := dynamic.Resolve(r.Ref, &resolved, config, reader); err != nil {
			return err
		}
		r.Value = resolved.Value
	}
	return nil
}
