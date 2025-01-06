package asyncapi3test

import (
	"mokapi/providers/asyncapi3"
)

type ConfigOptions func(c *asyncapi3.Config)

func NewConfig(opts ...ConfigOptions) *asyncapi3.Config {
	c := &asyncapi3.Config{
		Version:            "2.0.0",
		Info:               asyncapi3.Info{Name: "test", Version: "1.0"},
		Servers:            map[string]*asyncapi3.ServerRef{},
		DefaultContentType: "application/json",
	}
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

func WithServer(name, protocol, host string, opts ...ServerOptions) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Servers == nil {
			c.Servers = make(map[string]*asyncapi3.ServerRef)
		}

		s := &asyncapi3.Server{
			Host:     host,
			Protocol: protocol,
		}
		for _, opt := range opts {
			opt(s)
		}

		c.Servers[name] = &asyncapi3.ServerRef{Value: s}
	}
}

func WithChannel(name string, opts ...ChannelOptions) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Channels == nil {
			c.Channels = make(map[string]*asyncapi3.ChannelRef)
		}
		ch := NewChannel(opts...)
		c.Channels[name] = &asyncapi3.ChannelRef{Value: ch}
	}
}

func AddChannel(name string, ch *asyncapi3.Channel) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Channels == nil {
			c.Channels = make(map[string]*asyncapi3.ChannelRef)
		}
		c.Channels[name] = &asyncapi3.ChannelRef{Value: ch}
	}
}

func WithOperation(name string, opts ...OperationOptions) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Operations == nil {
			c.Operations = make(map[string]*asyncapi3.OperationRef)
		}
		op := NewOperation(opts...)
		switch op.Action {
		case "send", "receive":
		default:
			panic("no valid action set: expected send or receive")
		}
		c.Operations[name] = &asyncapi3.OperationRef{Value: op}
	}
}
