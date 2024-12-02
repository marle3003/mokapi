package store

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/kafka"
	"mokapi/media"
	"mokapi/schema/encoding"
	"mokapi/schema/json/schema"
)

type validator struct {
	payload     *schema.Ref
	contentType string
}

func newValidator(c *asyncApi.Channel) *validator {
	return &validator{
		payload:     getPayload(c),
		contentType: getContentType(c),
	}
}

func (v *validator) update(c *asyncApi.Channel) {
	v.payload = getPayload(c)
	v.contentType = getContentType(c)
}

func (v *validator) Payload(payload kafka.Bytes) error {
	if len(v.contentType) == 0 || v.payload == nil {
		return nil
	}

	_, err := encoding.DecodeFrom(payload, encoding.WithContentType(media.ParseContentType(v.contentType)), encoding.WithSchema(v.payload))
	return err
}

func getPayload(c *asyncApi.Channel) *schema.Ref {
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
