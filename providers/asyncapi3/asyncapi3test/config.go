package asyncapi3test

import (
	"mokapi/config/dynamic"
	"mokapi/providers/asyncapi3"
)

type ConfigOptions func(c *asyncapi3.Config)

func NewConfig(opts ...ConfigOptions) *asyncapi3.Config {
	c := &asyncapi3.Config{
		Version: "2.0.0",
		Info:    asyncapi3.Info{Name: "test", Version: "1.0"},
		Servers: map[string]*asyncapi3.ServerRef{}}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithTitle(title string) ConfigOptions {
	return func(c *asyncapi3.Config) {
		c.Info.Name = title
	}
}

func WithInfo(title, description, version string) ConfigOptions {
	return func(c *asyncapi3.Config) {
		c.Info.Name = title
		c.Info.Description = description
		c.Info.Version = version
	}
}

func WithInfoExt(termsOfService, licenseName, licenseUrl string) ConfigOptions {
	return func(c *asyncapi3.Config) {
		c.Info.TermsOfService = termsOfService
		c.Info.License = &asyncapi3.License{
			Name: licenseName,
			Url:  licenseUrl,
		}
	}
}

func WithContact(name, url, mail string) ConfigOptions {
	return func(c *asyncapi3.Config) {
		c.Info.Contact = &asyncapi3.Contact{
			Name:  name,
			Url:   url,
			Email: mail,
		}
	}
}

func WithChannelDescription(description string) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Description = description
	}
}

func AssignToServer(ref string) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Servers = append(c.Servers, &asyncapi3.ServerRef{Reference: dynamic.Reference{Ref: ref}})
	}
}

func WithTopicBinding(bindings asyncapi3.TopicBindings) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Bindings.Kafka = bindings
	}
}

type ServerOptions func(s *asyncapi3.Server)

func WithServer(name, protocol, host string, opts ...ServerOptions) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Servers == nil {
			c.Servers = make(map[string]*asyncapi3.ServerRef)
		}

		s := &asyncapi3.Server{
			Host:     host,
			Protocol: protocol,
		}
		s.Bindings.Kafka.Config = make(map[string]string)
		for _, opt := range opts {
			opt(s)
		}

		c.Servers[name] = &asyncapi3.ServerRef{Value: s}
	}
}

func WithServerDescription(description string) ServerOptions {
	return func(s *asyncapi3.Server) {
		s.Description = description
	}
}

func WithServerTags(tags ...asyncapi3.Tag) ServerOptions {
	return func(s *asyncapi3.Server) {
		for _, tag := range tags {
			s.Tags = append(s.Tags, &asyncapi3.TagRef{Value: &tag})
		}
	}
}

func WithKafkaBinding(key, value string) ServerOptions {
	return func(s *asyncapi3.Server) {
		s.Bindings.Kafka.Config[key] = value
	}
}
