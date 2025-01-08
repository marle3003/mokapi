package asyncapi3test

import (
	"mokapi/providers/asyncapi3"
	"mokapi/schema/json/schema"
)

type MessageOptions func(m *asyncapi3.Message)

func NewMessage(opts ...MessageOptions) *asyncapi3.Message {
	m := &asyncapi3.Message{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func WithPayload(s *schema.Schema) MessageOptions {
	return func(m *asyncapi3.Message) {
		m.Payload = &asyncapi3.SchemaRef{Value: &asyncapi3.MultiSchemaFormat{Schema: s}}
	}
}

func WithPayloadMulti(format string, schema asyncapi3.Schema) MessageOptions {
	return func(m *asyncapi3.Message) {
		m.Payload = &asyncapi3.SchemaRef{Value: &asyncapi3.MultiSchemaFormat{
			Format: format,
			Schema: schema,
		}}
	}
}

func WithContentType(s string) MessageOptions {
	return func(m *asyncapi3.Message) {
		m.ContentType = s
	}
}

func WithKey(s *schema.Schema) MessageOptions {
	return func(m *asyncapi3.Message) {
		m.Bindings.Kafka.Key = &asyncapi3.SchemaRef{Value: &asyncapi3.MultiSchemaFormat{Schema: s}}
	}
}

func WithMessageInfo(name, title, summary, description string) MessageOptions {
	return func(m *asyncapi3.Message) {
		m.Name = name
		m.Title = title
		m.Summary = summary
		m.Description = description
	}
}

func WithMessageTrait(trait *asyncapi3.MessageTrait) MessageOptions {
	return func(m *asyncapi3.Message) {
		m.Traits = append(m.Traits, &asyncapi3.MessageTraitRef{Value: trait})
	}
}

func WithKafkaMessageBinding(b asyncapi3.KafkaMessageBinding) MessageOptions {
	return func(m *asyncapi3.Message) {
		m.Bindings.Kafka = b
	}
}
