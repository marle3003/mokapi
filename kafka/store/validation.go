package store

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/openapi"
	"mokapi/kafka/protocol"
	"mokapi/models/media"
	"mokapi/providers/encoding"
)

type validator struct {
	payload     *openapi.SchemaRef
	contentType string
}

func newValidator(c *asyncApi.Channel) *validator {
	return &validator{
		payload:     getPayload(c),
		contentType: getContentType(c),
	}
}

func (v *validator) Payload(payload protocol.Bytes) error {
	if len(v.contentType) == 0 || v.payload == nil {
		return nil
	}

	_, err := encoding.ParseFrom(payload, media.ParseContentType(v.contentType), v.payload)
	return err
}

func getPayload(c *asyncApi.Channel) *openapi.SchemaRef {
	if c.Publish == nil ||
		c.Publish.Message == nil ||
		c.Publish.Message.Value == nil {
		return nil
	}
	return c.Publish.Message.Value.Payload
}

func getContentType(c *asyncApi.Channel) string {
	if c.Publish == nil ||
		c.Publish.Message == nil ||
		c.Publish.Message.Value == nil {
		return ""
	}
	return c.Publish.Message.Value.ContentType
}
