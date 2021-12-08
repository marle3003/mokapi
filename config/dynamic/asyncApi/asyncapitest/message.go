package asyncapitest

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/openapi"
)

type MessageOptions func(m *asyncApi.Message)

func NewMessage(opts ...MessageOptions) *asyncApi.Message {
	m := &asyncApi.Message{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func WithPayload(schema *openapi.Schema) MessageOptions {
	return func(m *asyncApi.Message) {
		m.Payload = &openapi.SchemaRef{Value: schema}
	}
}
