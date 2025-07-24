package asyncapitest

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/providers/asyncapi3"
	"mokapi/schema/json/schema"
)

type MessageOptions func(m *asyncApi.Message)

func NewMessage(opts ...MessageOptions) *asyncApi.Message {
	m := &asyncApi.Message{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func WithPayload(s *schema.Schema) MessageOptions {
	return func(m *asyncApi.Message) {
		m.Payload = &asyncapi3.SchemaRef{Value: &asyncapi3.MultiSchemaFormat{Schema: s}}
	}
}

func WithContentType(s string) MessageOptions {
	return func(m *asyncApi.Message) {
		m.ContentType = s
	}
}

func WithMessageInfo(name, title, summary, description string) MessageOptions {
	return func(m *asyncApi.Message) {
		m.Name = name
		m.Title = title
		m.Summary = summary
		m.Description = description
	}
}

func WithMessageId(messageId string) MessageOptions {
	return func(m *asyncApi.Message) {
		m.MessageId = messageId
	}
}
