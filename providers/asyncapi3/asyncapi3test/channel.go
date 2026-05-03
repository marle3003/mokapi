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

func UseMessage(name string, msg *asyncapi3.MessageRef) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		if c.Messages == nil {
			c.Messages = make(map[string]*asyncapi3.MessageRef)
		}
		c.Messages[name] = msg
	}
}

func WithKafkaChannelBinding(bindings asyncapi3.TopicBindings) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Bindings.Kafka = bindings
	}
}

func WithChannelAddress(address string) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Address = address
	}
}

func WithChannelName(name string) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Name = name
	}
}

func WithChannelTitle(title string) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Title = title
	}
}

func WithChannelSummary(summary string) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Summary = summary
	}
}

func WithChannelDescription(desc string) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Description = desc
	}
}

func AssignToServer(ref string) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Servers = append(c.Servers, &asyncapi3.ServerRef{Reference: dynamic.Reference[*asyncapi3.ServerRef]{Ref: ref}})
	}
}

func WithChannelTag(name, description string) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		c.Tags = append(c.Tags, &asyncapi3.TagRef{
			Value: &asyncapi3.Tag{
				Name:        name,
				Description: description,
			},
		})
	}
}

func WithParameter(name string, param *asyncapi3.Parameter) ChannelOptions {
	return func(c *asyncapi3.Channel) {
		if c.Parameters == nil {
			c.Parameters = map[string]*asyncapi3.ParameterRef{}
		}
		c.Parameters[name] = &asyncapi3.ParameterRef{
			Value: param,
		}
	}
}
