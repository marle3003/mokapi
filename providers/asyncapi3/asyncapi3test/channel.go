package asyncapi3test

import (
	"mokapi/config/dynamic"
	"mokapi/providers/asyncapi3"
)

type ChannelOptions func(c *asyncapi3.Channel)

func NewChannel(opts ...ChannelOptions) *asyncapi3.Channel {
	ch := &asyncapi3.Channel{}
	// default enable validation
	ch.Bindings.Kafka.ValueSchemaValidation = true
	ch.Bindings.Kafka.Partitions = 1
	for _, opt := range opts {
		opt(ch)
	}
	return ch
}

func WithMessage(name string, opts ...MessageOptions) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		msg := NewMessage(opts...)
		if c.Messages == nil {
			c.Messages = make(map[string]*asyncapi3.MessageRef)
		}
		c.Messages[name] = &asyncapi3.MessageRef{Value: msg}
	}
}

func WithKafkaChannelBinding(bindings asyncapi3.TopicBindings) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Bindings.Kafka = bindings
	}
}

func WithChannelDescription(desc string) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Description = desc
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
