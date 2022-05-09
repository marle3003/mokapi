package asyncapitest

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/openapi/schema"
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
		m.Payload = &schema.Ref{Value: s}
	}
}

func WithContentType(s string) MessageOptions {
	return func(m *asyncApi.Message) {
		m.ContentType = s
	}
}

func WithKey(s *schema.Schema) MessageOptions {
	return func(m *asyncApi.Message) {
		m.Bindings.Kafka.Key = &schema.Ref{Value: s}
	}
}
