package asyncapitest

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/schema/json/schema"
)

type ConfigOptions func(c *asyncApi.Config)

func NewConfig(opts ...ConfigOptions) *asyncApi.Config {
	c := &asyncApi.Config{
		AsyncApi: "2.0.0",
		Info:     asyncApi.Info{Name: "test", Version: "1.0"},
		Servers:  map[string]*asyncApi.ServerRef{}}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithTitle(title string) ConfigOptions {
	return func(c *asyncApi.Config) {
		c.Info.Name = title
	}
}

func WithInfo(title, description, version string) ConfigOptions {
	return func(c *asyncApi.Config) {
		c.Info.Name = title
		c.Info.Description = description
		c.Info.Version = version
	}
}

func WithInfoExt(termsOfService, licenseName, licenseUrl string) ConfigOptions {
	return func(c *asyncApi.Config) {
		c.Info.TermsOfService = termsOfService
		c.Info.License = &asyncApi.License{
			Name: licenseName,
			Url:  licenseUrl,
		}
	}
}

func WithContact(name, url, mail string) ConfigOptions {
	return func(c *asyncApi.Config) {
		c.Info.Contact = &asyncApi.Contact{
			Name:  name,
			Url:   url,
			Email: mail,
		}
	}
}

func WithChannel(name string, opts ...ChannelOptions) ConfigOptions {
	return func(c *asyncApi.Config) {
		if c.Channels == nil {
			c.Channels = make(map[string]*asyncApi.ChannelRef)
		}
		ch := NewChannel(opts...)
		c.Channels[name] = &asyncApi.ChannelRef{Value: ch}
	}
}

func WithSchemas(name string, s *schema.Schema) ConfigOptions {
	return func(c *asyncApi.Config) {
		if c.Components == nil {
			c.Components = &asyncApi.Components{}
		}
		if c.Components.Schemas == nil {
			c.Components.Schemas = &schema.Schemas{}
		}
		c.Components.Schemas.Set(name, &schema.Ref{Value: s})
	}
}

func WithMessages(name string, message *asyncApi.Message) ConfigOptions {
	return func(c *asyncApi.Config) {
		if c.Components == nil {
			c.Components = &asyncApi.Components{}
		}
		if c.Components.Messages == nil {
			c.Components.Messages = map[string]*asyncApi.Message{}
		}
		c.Components.Messages[name] = message
	}
}

func WithChannelDescription(description string) ChannelOptions {
	return func(c *asyncApi.Channel) {
		c.Description = description
	}
}

func AssignToServer(server string) ChannelOptions {
	return func(c *asyncApi.Channel) {
		c.Servers = append(c.Servers, server)
	}
}

func WithTopicBinding(bindings asyncApi.TopicBindings) ChannelOptions {
	return func(c *asyncApi.Channel) {
		c.Bindings.Kafka = bindings
	}
}

type ServerOptions func(s *asyncApi.Server)

func WithServer(name, protocol, url string, opts ...ServerOptions) ConfigOptions {
	return func(c *asyncApi.Config) {
		if c.Servers == nil {
			c.Servers = make(map[string]*asyncApi.ServerRef)
		}

		s := &asyncApi.Server{
			Url:      url,
			Protocol: protocol,
		}
		s.Bindings.Kafka.Config = make(map[string]string)
		for _, opt := range opts {
			opt(s)
		}

		c.Servers[name] = &asyncApi.ServerRef{Value: s}
	}
}

func WithServerDescription(description string) ServerOptions {
	return func(s *asyncApi.Server) {
		s.Description = description
	}
}

func WithServerTags(tags ...asyncApi.ServerTag) ServerOptions {
	return func(s *asyncApi.Server) {
		s.Tags = append(s.Tags, tags...)
	}
}

func WithKafkaBinding(key, value string) ServerOptions {
	return func(s *asyncApi.Server) {
		s.Bindings.Kafka.Config[key] = value
	}
}
