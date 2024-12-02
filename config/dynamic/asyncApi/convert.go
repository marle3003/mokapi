package asyncApi

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/providers/asyncapi3"
	"mokapi/schema/json/schema"
	"net/url"
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

	return target, nil
}

func convertChannels(cfg *asyncapi3.Config, channels map[string]*ChannelRef) error {
	for name, orig := range channels {
		target := &asyncapi3.ChannelRef{Reference: dynamic.Reference{Ref: orig.Ref}}
		if orig.Value != nil {
			target.Value = &asyncapi3.Channel{
				Address:     name,
				Summary:     "",
				Description: orig.Value.Description,
				Bindings:    convertChannelBinding(orig.Value.Bindings),
			}

			for _, server := range orig.Value.Servers {
				target.Value.Servers = append(
					target.Value.Servers,
					&asyncapi3.ServerRef{Reference: dynamic.Reference{
						Ref: fmt.Sprintf("#/servers/%s", server),
					}})
			}

			if err := convertParameters(target.Value, orig.Value.Parameters); err != nil {
				return err
			}

			convertOperation(target.Value, orig.Value.Publish, "publish")
			convertOperation(target.Value, orig.Value.Subscribe, "subscribe")
		}

		if cfg.Channels == nil {
			cfg.Channels = map[string]*asyncapi3.ChannelRef{}
		}

		cfg.Channels[name] = target
	}

	return nil
}

func convertOperation(channel *asyncapi3.Channel, op *Operation, opName string) {
	if op == nil {
		return
	}

	msg := convertMessage(op.Message)
	var msgId string
	if len(op.OperationId) > 0 {
		msgId = fmt.Sprintf("%v.message", op.OperationId)
	} else {
		msgId = fmt.Sprintf("%v.message", opName)
	}

	if channel.Messages == nil {
		channel.Messages = map[string]*asyncapi3.MessageRef{}
	}

	channel.Messages[msgId] = msg
}

func convertMessage(msg *MessageRef) *asyncapi3.MessageRef {
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
		for _, trait := range msg.Value.Traits {
			if trait.Value == nil {
				continue
			}
			target.Value.Traits = append(target.Value.Traits, convertMessageTrait(trait.Value))
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
	return &asyncapi3.SchemaRef{Reference: dynamic.Reference{Ref: s.Ref}, Value: s}
}

func convertParameters(channel *asyncapi3.Channel, params map[string]*ParameterRef) error {
	for name, orig := range params {
		target := &asyncapi3.ParameterRef{Reference: dynamic.Reference{Ref: orig.Ref}}
		if orig.Value != nil {
			target.Value = &asyncapi3.Parameter{
				Description: orig.Value.Description,
				Location:    orig.Value.Location,
			}

			if orig.Value.Schema != nil {
				for _, enum := range orig.Value.Schema.Enum {
					if s, ok := enum.(string); !ok {
						return fmt.Errorf("unable to convert parameter %v: only string enum values supported: %v", name, enum)
					} else {
						target.Value.Enum = append(target.Value.Enum, s)
					}
				}

				if orig.Value.Schema.Default != nil {
					if s, ok := orig.Value.Schema.Default.(string); !ok {
						return fmt.Errorf("unable to convert parameter %v: only string default value supported: %v", name, orig.Value.Schema.Default)
					} else {
						target.Value.Default = s
					}
				}

				for _, example := range orig.Value.Schema.Examples {
					if s, ok := example.(string); !ok {
						return fmt.Errorf("unable to convert parameter %v: only string example values supported: %v", name, example)
					} else {
						target.Value.Examples = append(target.Value.Enum, s)
					}
				}
			}
		}

		if channel.Parameters == nil {
			channel.Parameters = map[string]*asyncapi3.ParameterRef{}
		}

		channel.Parameters[name] = target
	}
	return nil
}

func convertServers(cfg *asyncapi3.Config, servers map[string]*ServerRef) {
	for name, orig := range servers {
		target := &asyncapi3.ServerRef{Reference: dynamic.Reference{Ref: orig.Ref}}
		if orig.Value != nil {
			target.Value = &asyncapi3.Server{
				Protocol:        orig.Value.Protocol,
				Description:     orig.Value.Description,
				ProtocolVersion: orig.Value.ProtocolVersion,
				Bindings:        convertServerBinding(orig.Value.Bindings),
			}

			protocol, host, path := resolveServerUrl(orig.Value.Url)
			if target.Value.Protocol == "" {
				target.Value.Protocol = protocol
			}
			target.Value.Host = host
			target.Value.Pathname = path
		}

		if cfg.Servers == nil {
			cfg.Servers = map[string]*asyncapi3.ServerRef{}
		}

		cfg.Servers[name] = target
	}
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
	return asyncapi3.MessageBinding{Kafka: asyncapi3.KafkaMessageBinding{Key: convertSchema(b.Kafka.Key, "")}}
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
