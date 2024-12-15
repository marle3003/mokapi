package store

import (
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	"mokapi/schema/encoding"
	"mokapi/schema/json/schema"
)

type validator struct {
	validators []recordValidator
}

type recordValidator interface {
	Validate(record *kafka.Record) error
}

func newValidator(c *asyncapi3.Channel) *validator {
	v := &validator{}
	v.update(c)
	return v
}

func (v *validator) update(c *asyncapi3.Channel) {
	v.validators = nil

	for id, msg := range c.Messages {
		if msg.Value == nil || msg.Value.Payload == nil {
			continue
		}
		v.validators = append(v.validators, newMessageValidator(id, msg.Value))
	}
}

func (v *validator) Validate(record *kafka.Record) error {
	var err error
	for _, v := range v.validators {
		err = v.Validate(record)
		if err == nil {
			return nil
		}
	}
	return err
}

type messageValidator struct {
	messageId string
	payload   schemaValidator
}

func newMessageValidator(messageId string, msg *asyncapi3.Message) *messageValidator {
	v := &messageValidator{messageId: messageId}

	switch s := msg.Payload.Value.(type) {
	case *schema.Ref:
		v.payload = &jsonSchemaValidator{
			schema:      s,
			contentType: msg.ContentType,
		}
	default:
		log.Errorf("unsupported payload type: %T", msg.Payload.Value)
	}
	return v
}

func (mv *messageValidator) Validate(record *kafka.Record) error {
	if mv.payload != nil {
		if err := mv.payload.Validate(record.Value); err != nil {
			return err
		}
	}

	record.Headers = append(record.Headers, kafka.RecordHeader{
		Key:   "x-specification-message-id",
		Value: []byte(mv.messageId),
	})

	return nil
}

type schemaValidator interface {
	Validate(record kafka.Bytes) error
}

type jsonSchemaValidator struct {
	schema      *schema.Ref
	contentType string
}

func (v *jsonSchemaValidator) Validate(data kafka.Bytes) error {
	if len(v.contentType) == 0 || v.schema == nil {
		return nil
	}

	_, err := encoding.DecodeFrom(data, encoding.WithContentType(media.ParseContentType(v.contentType)), encoding.WithSchema(v.schema))
	return err
}