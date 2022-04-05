package asyncapitest

import "mokapi/config/dynamic/asyncApi"

type ConfigOptions func(c *asyncApi.Config)

func NewConfig(opts ...ConfigOptions) *asyncApi.Config {
	c := &asyncApi.Config{
		AsyncApi: "2.0.0",
		Info:     asyncApi.Info{Name: "test", Version: "1.0"},
		Servers:  map[string]asyncApi.Server{}}
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

func WithChannel(name string, opts ...ChannelOptions) ConfigOptions {
	return func(c *asyncApi.Config) {
		if c.Channels == nil {
			c.Channels = make(map[string]*asyncApi.ChannelRef)
		}
		ch := NewChannel(opts...)
		c.Channels[name] = &asyncApi.ChannelRef{Value: ch}
	}
}

func WithChannelBinding(key, value string) ChannelOptions {
	return func(c *asyncApi.Channel) {
		c.Bindings.Kafka.Config[key] = value
	}
}

type ServerOptions func(s asyncApi.Server)

func WithServer(name, protocol, url string, opts ...ServerOptions) ConfigOptions {
	return func(c *asyncApi.Config) {
		if c.Servers == nil {
			c.Servers = make(map[string]asyncApi.Server)
		}

		s := asyncApi.Server{
			Url:      url,
			Protocol: protocol,
		}
		s.Bindings.Kafka.Config = make(map[string]string)
		for _, opt := range opts {
			opt(s)
		}

		c.Servers[name] = s
	}
}

func WithKafka(key, value string) ServerOptions {
	return func(s asyncApi.Server) {
		s.Bindings.Kafka.Config[key] = value
	}
}
