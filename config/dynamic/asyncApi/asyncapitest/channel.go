package asyncapitest

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/openapi/schema"
)

type ChannelOptions func(c *asyncApi.Channel)

func NewChannel(opts ...ChannelOptions) *asyncApi.Channel {
	ch := &asyncApi.Channel{}
	ch.Bindings.Kafka.Config = make(map[string]string)
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

func WithOperationBinding(groupId *schema.Schema) OperationOptions {
	return func(o *asyncApi.Operation) {
		o.Bindings.Kafka.GroupId = groupId
	}
}

func WithChannelKafka(key, value string) ChannelOptions {
	return func(c *asyncApi.Channel) {
		c.Bindings.Kafka.Config[key] = value
	}
}
