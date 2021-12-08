package asyncapitest

import "mokapi/config/dynamic/asyncApi"

type ConfigOptions func(c *asyncApi.Config)

func NewConfig(opts ...ConfigOptions) *asyncApi.Config {
	c := &asyncApi.Config{}
	for _, opt := range opts {
		opt(c)
	}
	return c
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
