package asyncapitest

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka"
)

type ChannelOptions func(c *asyncApi.Channel)

func NewChannel(opts ...ChannelOptions) *asyncApi.Channel {
	ch := &asyncApi.Channel{}
	for _, opt := range opts {
		opt(ch)
	}
	return ch
}

type OperationOptions func(o *asyncApi.Operation)

func WithSubscribeAndPublish(opts ...OperationOptions) ChannelOptions {
	return func(c *asyncApi.Channel) {
		c.Subscribe = &asyncApi.Operation{}
		for _, opt := range opts {
			opt(c.Subscribe)
		}
		c.Publish = &asyncApi.Operation{}
		for _, opt := range opts {
			opt(c.Publish)
		}
	}
}

func WithSubscribe(opts ...OperationOptions) ChannelOptions {
	return func(c *asyncApi.Channel) {
		c.Subscribe = &asyncApi.Operation{}
		for _, opt := range opts {
			opt(c.Subscribe)
		}
	}
}

func WithPublish(opts ...OperationOptions) ChannelOptions {
	return func(c *asyncApi.Channel) {
		c.Publish = &asyncApi.Operation{}
		for _, opt := range opts {
			opt(c.Publish)
		}
	}
}

func WithMessage(opts ...MessageOptions) OperationOptions {
	return func(o *asyncApi.Operation) {
		if o.Message == nil {
			o.Message = &asyncApi.MessageRef{Value: &asyncApi.Message{}}
		}
		for _, opt := range opts {
			opt(o.Message.Value)
		}
	}
}

func WithOperationInfo(id, summary, description string) OperationOptions {
	return func(o *asyncApi.Operation) {
		o.OperationId = id
		o.Summary = summary
		o.Description = description
	}
}

func WithOperationBinding(b kafka.Operation) OperationOptions {
	return func(o *asyncApi.Operation) {
		o.Bindings.Kafka = b
	}
}

func WithChannelKafka(bindings kafka.TopicBindings) ChannelOptions {
	return func(c *asyncApi.Channel) {
		c.Bindings.Kafka = bindings
	}
}
