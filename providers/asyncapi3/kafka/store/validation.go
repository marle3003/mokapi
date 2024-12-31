package store

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
)

type validator struct {
	validators []recordValidator
}

type recordValidator interface {
	Validate(record *kafka.Record) (interface{}, interface{}, error)
}

func newValidator(c *asyncapi3.Channel) *validator {
	v := &validator{}

	for id, msg := range c.Messages {
		if msg.Value == nil || msg.Value.Payload == nil {
			continue
		}
		v.validators = append(v.validators, newMessageValidator(id, msg.Value))
	}

	return v
}

func (v *validator) Validate(record *kafka.Record) (key interface{}, payload interface{}, err error) {
	if v == nil {
		return record.Key, record.Value, nil
	}

	for _, val := range v.validators {
		key, payload, err = val.Validate(record)
		if err == nil {
			return
		}
	}
	return record.Key, record.Value, err
}

type messageValidator struct {
	messageId string
	key       *schemaValidator
	payload   *schemaValidator
}

func newMessageValidator(messageId string, msg *asyncapi3.Message) *messageValidator {
	v := &messageValidator{messageId: messageId}

	var msgParser encoding.Parser
	switch s := msg.Payload.Value.Schema.(type) {
	case *schema.Ref:
		msgParser = &parser.Parser{Schema: s}
	case *avro.Schema:
		msgParser = &avro.Parser{Schema: s}
	default:
		log.Errorf("unsupported payload type: %T", msg.Payload.Value)
	}

	if msgParser != nil {
		v.payload = &schemaValidator{
			parser:      msgParser,
			contentType: msg.ContentType,
		}
	}

	if msg.Bindings.Kafka.Key != nil {
		var keyParser encoding.Parser
		switch s := msg.Bindings.Kafka.Key.Value.Schema.(type) {
		case *schema.Ref:
			keyParser = &parser.Parser{Schema: s}
		case *avro.Schema:
			keyParser = &avro.Parser{Schema: s}
		default:
			log.Errorf("unsupported key type: %T", msg.Bindings.Kafka.Key.Value)
		}

		if keyParser != nil {
			v.key = &schemaValidator{
				parser: keyParser,
			}
		}
	}

	return v
}

func (mv *messageValidator) Validate(record *kafka.Record) (key interface{}, payload interface{}, err error) {
	if mv.payload != nil {
		if payload, err = mv.payload.Validate(record.Value); err != nil {
			err = fmt.Errorf("invalid message: %w", err)
			return
		}
	}

	if mv.key != nil {
		if key, err = mv.key.Validate(record.Key); err != nil {
			err = fmt.Errorf("invalid key: %w", err)
			return
		}
	} else {
		key = kafka.BytesToString(record.Key)
	}

	record.Headers = append(record.Headers, kafka.RecordHeader{
		Key:   "x-specification-message-id",
		Value: []byte(mv.messageId),
	})

	return
}

type schemaValidator struct {
	parser      encoding.Parser
	contentType string
}

func (v *schemaValidator) Validate(data kafka.Bytes) (interface{}, error) {
	return encoding.DecodeFrom(data, encoding.WithContentType(media.ParseContentType(v.contentType)), encoding.WithParser(v.parser))
}
