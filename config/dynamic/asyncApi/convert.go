package asyncApi

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/providers/openapi/schema"
	"net/url"
	"strings"
)

func (c *Config) Convert() (*Config3, error) {
	target := &Config3{
		Version:            "3.0.0",
		Id:                 c.Id,
		DefaultContentType: c.DefaultContentType,
	}
	target.Info = c.Info

	convertServers(target, c.Servers)
	if err := convertChannels(target, c.Channels); err != nil {
		return nil, err
	}

	return target, nil
}

func convertChannels(cfg *Config3, channels map[string]*ChannelRef) error {
	for name, orig := range channels {
		target := &Channel3Ref{Reference: dynamic.Reference{Ref: orig.Ref}}
		if orig.Value != nil {
			target.Value = &Channel3{
				Address:     name,
				Summary:     "",
				Description: orig.Value.Description,
				Bindings:    orig.Value.Bindings,
			}

			for _, server := range orig.Value.Servers {
				target.Value.Servers = append(
					target.Value.Servers,
					&Server3Ref{Reference: dynamic.Reference{
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
			cfg.Channels = map[string]*Channel3Ref{}
		}

		cfg.Channels[name] = target
	}

	return nil
}

func convertOperation(channel *Channel3, op *Operation, opName string) {
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
		channel.Messages = map[string]*Message3Ref{}
	}

	channel.Messages[msgId] = msg
}

func convertMessage(msg *MessageRef) *Message3Ref {
	target := &Message3Ref{Reference: dynamic.Reference{Ref: msg.Ref}}
	if msg.Value != nil {
		target.Value = &Message3{
			Title:         msg.Value.Title,
			Name:          msg.Value.Name,
			Summary:       msg.Value.Summary,
			Description:   msg.Value.Description,
			CorrelationId: msg.Value.CorrelationId,
			ContentType:   msg.Value.ContentType,
			Payload:       nil,
			Bindings:      msg.Value.Bindings,
			Traits:        msg.Value.Traits,
			ExternalDocs:  nil,
		}

		if msg.Value.Payload != nil {
			target.Value.Payload = convertSchema(msg.Value.Payload, msg.Value.SchemaFormat)
		}
		if msg.Value.Headers != nil {
			target.Value.Headers = convertSchema(msg.Value.Headers, "")
		}
	}
	return target
}

func convertSchema(s *schema.Ref, schemaFormat string) *SchemaRef {
	target := &SchemaRef{Reference: dynamic.Reference{Ref: s.Ref}}

	if schemaFormat == "" {
		target.Value = schema.ConvertToJsonSchema(s).Value
	} else {
		target.Value = &MultiSchemaFormat{
			Format: schemaFormat,
			Schema: s,
		}
	}
	return target
}

func convertParameters(channel *Channel3, params map[string]*ParameterRef) error {
	for name, orig := range params {
		target := &Parameter3Ref{Reference: dynamic.Reference{Ref: orig.Ref}}
		if orig.Value != nil {
			target.Value = &Parameter3{
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
			channel.Parameters = map[string]*Parameter3Ref{}
		}

		channel.Parameters[name] = target
	}
	return nil
}

func convertServers(cfg *Config3, servers map[string]*ServerRef) {
	for name, orig := range servers {
		target := &Server3Ref{Reference: dynamic.Reference{Ref: orig.Ref}}
		if orig.Value != nil {
			target.Value = &Server3{
				Protocol:        orig.Value.Protocol,
				Description:     orig.Value.Description,
				ProtocolVersion: orig.Value.ProtocolVersion,
				Variables:       orig.Value.Variables,
				Bindings:        orig.Value.Bindings,
			}

			protocol, host, path := resolveServerUrl(orig.Value.Url)
			if target.Value.Protocol == "" {
				target.Value.Protocol = protocol
			}
			target.Value.Host = host
			target.Value.Pathname = path
		}

		if cfg.Servers == nil {
			cfg.Servers = map[string]*Server3Ref{}
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
