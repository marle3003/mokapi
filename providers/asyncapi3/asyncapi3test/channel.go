package asyncapi3test

import (
	"mokapi/config/dynamic"
	"mokapi/providers/asyncapi3"
)

type ChannelOptions func(c *asyncapi3.Channel)

func NewChannel(opts ...ChannelOptions) *asyncapi3.Channel {
	ch := &asyncapi3.Channel{}
	for _, opt := range opts {
		opt(ch)
	}
	return ch
}

type OperationOptions func(o *asyncapi3.Operation)

func WithMessage(name string, opts ...MessageOptions) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		msg := NewMessage(opts...)
		if c.Messages == nil {
			c.Messages = make(map[string]*asyncapi3.MessageRef)
		}
		c.Messages[name] = &asyncapi3.MessageRef{Value: msg}
	}
}

func WithOperationMessage(ref string) OperationOptions {
	return func(o *asyncapi3.Operation) {
		o.Messages = append(o.Messages, asyncapi3.MessageRef{Reference: dynamic.Reference{Ref: ref}})
	}
}

func WithOperationInfo(summary, description string) OperationOptions {
	return func(o *asyncapi3.Operation) {
		o.Summary = summary
		o.Description = description
	}
}

func WithOperationBinding(b asyncapi3.KafkaOperation) OperationOptions {
	return func(o *asyncapi3.Operation) {
		o.Bindings.Kafka = b
	}
}

func WithChannelKafka(bindings asyncapi3.TopicBindings) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Bindings.Kafka = bindings
	}
}
