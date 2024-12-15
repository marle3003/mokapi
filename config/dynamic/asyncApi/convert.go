package asyncApi

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/providers/asyncapi3"
	"mokapi/schema/json/schema"
	"net/url"
	"path"
	"strings"
)

func (c *Config) Convert() (*asyncapi3.Config, error) {
	target := &asyncapi3.Config{
		Version:            "3.0.0",
		Id:                 c.Id,
		DefaultContentType: c.DefaultContentType,
	}
	target.Info = asyncapi3.Info{
		Name:           c.Info.Name,
		Description:    c.Info.Description,
		Version:        c.Info.Version,
		TermsOfService: c.Info.TermsOfService,
	}

	if c.Info.Contact != nil {
		target.Info.Contact = &asyncapi3.Contact{
			Name:  c.Info.Contact.Name,
			Url:   c.Info.Contact.Url,
			Email: c.Info.Contact.Email,
		}
	}

	if c.Info.License != nil {
		target.Info.License = &asyncapi3.License{
			Name: c.Info.License.Name,
			Url:  c.Info.License.Url,
		}
	}

	convertServers(target, c.Servers)
	if err := convertChannels(target, c.Channels); err != nil {
		return nil, err
	}

	if c.Components != nil {
		var err error
		target.Components, err = convertComponents(c.Components)
		if err != nil {
			return nil, err
		}
	}

	return target, nil
}

func convertChannels(cfg *asyncapi3.Config, channels map[string]*ChannelRef) error {
	if len(channels) == 0 {
		return nil
	}
	if cfg.Channels == nil {
		cfg.Channels = map[string]*asyncapi3.ChannelRef{}
	}

	for name, orig := range channels {
		if len(orig.Ref) > 0 {
			cfg.Channels[name] = &asyncapi3.ChannelRef{Reference: dynamic.Reference{Ref: orig.Ref}}
		}
		if orig.Value != nil {
			ch, err := convertChannel(name, orig.Value)
			if err != nil {
				return err
			}
			cfg.Channels[name] = ch
		}
	}

	return nil
}

func convertChannel(name string, c *Channel) (*asyncapi3.ChannelRef, error) {
	target := &asyncapi3.Channel{
		Address:     name,
		Summary:     "",
		Description: c.Description,
		Messages:    map[string]*asyncapi3.MessageRef{},
		Bindings:    convertChannelBinding(c.Bindings),
	}

	for _, server := range c.Servers {
		target.Servers = append(
			target.Servers,
			&asyncapi3.ServerRef{Reference: dynamic.Reference{
				Ref: fmt.Sprintf("#/servers/%s", server),
			}})
	}

	if err := convertParameters(target, c.Parameters); err != nil {
		return nil, err
	}

	if c.Publish != nil && c.Subscribe != nil && c.Publish.Message != nil && c.Subscribe.Message != nil {
		if c.Publish.Message.Ref == c.Subscribe.Message.Ref {
			msg := convertMessage(c.Publish.Message)
			addMessage(target, msg, "", "", c.Publish.Message.Ref)
		}
	} else {
		if c.Publish != nil {
			addMessage(target, convertMessage(c.Publish.Message), c.Publish.OperationId, c.Publish.Message.Ref, "publish")
		}
		if c.Subscribe != nil {
			addMessage(target, convertMessage(c.Subscribe.Message), c.Subscribe.OperationId, c.Subscribe.Message.Ref, "subscribe")
		}
	}

	return &asyncapi3.ChannelRef{Value: target}, nil
}

func addMessage(target *asyncapi3.Channel, msg *asyncapi3.MessageRef, opId, ref, opName string) {
	if msg == nil {
		return
	}

	var msgId string
	if len(opId) > 0 {
		msgId = opId
	} else if len(ref) > 0 {
		msgId = path.Base(ref)
	} else if len(opName) > 0 {
		msgId = opName
	}

	target.Messages[msgId] = msg
}

func convertMessage(msg *MessageRef) *asyncapi3.MessageRef {
	if msg == nil {
		return nil
	}

	target := &asyncapi3.MessageRef{Reference: dynamic.Reference{Ref: msg.Ref}}
	if msg.Value != nil {
		target.Value = &asyncapi3.Message{
			Title:        msg.Value.Title,
			Name:         msg.Value.Name,
			Summary:      msg.Value.Summary,
			Description:  msg.Value.Description,
			ContentType:  msg.Value.ContentType,
			Payload:      nil,
			Bindings:     convertMessageBinding(msg.Value.Bindings),
			ExternalDocs: nil,
		}

		if msg.Value.Payload != nil {
			target.Value.Payload = convertJsonSchema(msg.Value.Payload)
		}
		if msg.Value.Headers != nil {
			target.Value.Headers = convertJsonSchema(msg.Value.Headers)
		}
		for _, orig := range msg.Value.Traits {
			trait := &asyncapi3.MessageTraitRef{}
			if len(orig.Ref) > 0 {
				trait.Reference = dynamic.Reference{Ref: trait.Ref}
			}
			if orig.Value != nil {
				trait.Value = convertMessageTrait(orig.Value).Value
			}
			target.Value.Traits = append(target.Value.Traits, trait)
		}
	}
	return target
}

func convertMessageTrait(trait *MessageTrait) *asyncapi3.MessageTraitRef {
	target := &asyncapi3.MessageTrait{
		Name:        trait.Name,
		Title:       trait.Title,
		Summary:     trait.Summary,
		Description: trait.Description,
		ContentType: trait.ContentType,
		Bindings:    convertMessageBinding(trait.Bindings),
	}

	if trait.Headers != nil && trait.Headers.Value != nil {
		target.Headers = &asyncapi3.SchemaRef{
			Value: trait.Headers.Value,
		}
	}
	return &asyncapi3.MessageTraitRef{Value: target}
}

func convertSchema(s *SchemaRef, schemaFormat string) *asyncapi3.SchemaRef {
	if s == nil {
		return nil
	}
	target := &asyncapi3.SchemaRef{Reference: dynamic.Reference{Ref: s.Ref}}

	if schemaFormat == "" {
		target.Value = s.Value
	} else {
		target.Value = &MultiSchemaFormat{
			Format: schemaFormat,
			Schema: s,
		}
	}
	return target
}

func convertJsonSchema(s *schema.Ref) *asyncapi3.SchemaRef {
	return &asyncapi3.SchemaRef{Value: s}
}

func convertParameters(channel *asyncapi3.Channel, params map[string]*ParameterRef) error {
	if len(params) == 0 {
		return nil
	}
	if channel.Parameters == nil {
		channel.Parameters = map[string]*asyncapi3.ParameterRef{}
	}

	for name, orig := range params {
		if len(orig.Ref) > 0 {
			channel.Parameters[name] = &asyncapi3.ParameterRef{Reference: dynamic.Reference{Ref: orig.Ref}}
		}
		if orig.Value != nil {
			p, err := convertParameter(name, orig.Value)
			if err != nil {
				return err
			}
			channel.Parameters[name] = p
		}
	}
	return nil
}

func convertParameter(name string, p *Parameter) (*asyncapi3.ParameterRef, error) {
	target := &asyncapi3.Parameter{
		Description: p.Description,
		Location:    p.Location,
	}

	if p.Schema != nil {
		for _, enum := range p.Schema.Enum {
			if s, ok := enum.(string); !ok {
				return nil, fmt.Errorf("unable to convert parameter %v: only string enum values supported: %v", name, enum)
			} else {
				target.Enum = append(target.Enum, s)
			}
		}

		if p.Schema.Default != nil {
			if s, ok := p.Schema.Default.(string); !ok {
				return nil, fmt.Errorf("unable to convert parameter %v: only string default value supported: %v", name, p.Schema.Default)
			} else {
				target.Default = s
			}
		}

		for _, example := range p.Schema.Examples {
			if s, ok := example.(string); !ok {
				return nil, fmt.Errorf("unable to convert parameter %v: only string example values supported: %v", name, example)
			} else {
				target.Examples = append(target.Enum, s)
			}
		}
	}

	return &asyncapi3.ParameterRef{Value: target}, nil
}

func convertServers(cfg *asyncapi3.Config, servers map[string]*ServerRef) {
	if len(servers) == 0 {
		return
	}
	if cfg.Servers == nil {
		cfg.Servers = map[string]*asyncapi3.ServerRef{}
	}

	for name, orig := range servers {
		if len(orig.Ref) > 0 {
			cfg.Servers[name] = &asyncapi3.ServerRef{Reference: dynamic.Reference{Ref: orig.Ref}}
		}
		if orig.Value != nil {
			cfg.Servers[name] = convertServer(orig.Value)
		}
	}
}

func convertServer(s *Server) *asyncapi3.ServerRef {
	target := &asyncapi3.Server{
		Protocol:        s.Protocol,
		Description:     s.Description,
		ProtocolVersion: s.ProtocolVersion,
		Bindings:        convertServerBinding(s.Bindings),
	}

	protocol, host, path := resolveServerUrl(s.Url)
	if target.Protocol == "" {
		target.Protocol = protocol
	}
	target.Host = host
	target.Pathname = path

	return &asyncapi3.ServerRef{Value: target}
}

func resolveServerUrl(s string) (protocol, host, path string) {
	u, err := url.Parse(s)
	if err == nil && u.Scheme != "" && u.Host != "" {
		protocol = u.Scheme
		host = u.Host
	} else {
		host = s
	}
	split := strings.SplitN(host, "/", 1)
	if len(split) > 1 {
		host = split[0]
		path = split[1]
	}
	return
}

func convertServerBinding(b ServerBindings) asyncapi3.ServerBindings {
	return asyncapi3.ServerBindings{Kafka: asyncapi3.BrokerBindings{
		LogRetentionBytes:            b.Kafka.LogRetentionBytes,
		LogRetentionMs:               b.Kafka.LogRetentionMs,
		LogRetentionCheckIntervalMs:  b.Kafka.LogRetentionCheckIntervalMs,
		LogSegmentDeleteDelayMs:      b.Kafka.LogSegmentDeleteDelayMs,
		LogRollMs:                    b.Kafka.LogRollMs,
		LogSegmentBytes:              b.Kafka.LogSegmentBytes,
		GroupInitialRebalanceDelayMs: b.Kafka.GroupInitialRebalanceDelayMs,
		GroupMinSessionTimeoutMs:     b.Kafka.GroupMinSessionTimeoutMs,
	}}
}

func convertMessageBinding(b MessageBinding) asyncapi3.MessageBinding {
	return asyncapi3.MessageBinding{
		Kafka: asyncapi3.KafkaMessageBinding{
			Key: convertSchema(b.Kafka.Key, ""),
		},
	}
}

func convertChannelBinding(b ChannelBindings) asyncapi3.ChannelBindings {
	return asyncapi3.ChannelBindings{Kafka: asyncapi3.TopicBindings{
		Partitions:            b.Kafka.Partitions,
		RetentionBytes:        b.Kafka.RetentionBytes,
		RetentionMs:           b.Kafka.RetentionMs,
		SegmentBytes:          b.Kafka.SegmentBytes,
		SegmentMs:             b.Kafka.SegmentMs,
		ValueSchemaValidation: b.Kafka.ValueSchemaValidation,
	}}
}

func convertComponents(c *Components) (*asyncapi3.Components, error) {
	target := &asyncapi3.Components{}

	for name, server := range c.Servers {
		if target.Servers == nil {
			target.Servers = map[string]*asyncapi3.ServerRef{}
		}
		target.Servers[name] = convertServer(server)
	}

	for name, orig := range c.Channels {
		if target.Servers == nil {
			target.Channels = map[string]*asyncapi3.ChannelRef{}
		}
		ch, err := convertChannel(name, orig)
		if err != nil {
			return nil, err
		}
		target.Channels[name] = ch
	}

	for name, orig := range c.Messages {
		if target.Messages == nil {
			target.Messages = map[string]*asyncapi3.MessageRef{}
		}
		target.Messages[name] = convertMessage(&MessageRef{Value: orig})
	}

	if c.Schemas != nil {
		for it := c.Schemas.Iter(); it.Next(); {
			if target.Schemas == nil {
				target.Schemas = map[string]*asyncapi3.SchemaRef{}
			}
			target.Schemas[it.Key()] = convertSchema(&SchemaRef{Value: it.Value()}, "")
		}
	}

	for name, orig := range c.Parameters {
		if target.Parameters == nil {
			target.Parameters = map[string]*asyncapi3.ParameterRef{}
		}
		if len(orig.Ref) > 0 {
			target.Parameters[name] = &asyncapi3.ParameterRef{Reference: dynamic.Reference{Ref: orig.Ref}}
		}
		if orig.Value != nil {
			p, err := convertParameter(name, orig.Value)
			if err != nil {
				return nil, err
			}
			target.Parameters[name] = p
		}
	}

	for name, orig := range c.MessageTraits {
		if target.MessageTraits == nil {
			target.MessageTraits = map[string]*asyncapi3.MessageTraitRef{}
		}
		if len(orig.Ref) > 0 {
			target.MessageTraits[name] = &asyncapi3.MessageTraitRef{Reference: dynamic.Reference{Ref: orig.Ref}}
		}
		if orig.Value != nil {
			target.MessageTraits[name] = convertMessageTrait(orig.Value)
		}
	}

	return target, nil
}
