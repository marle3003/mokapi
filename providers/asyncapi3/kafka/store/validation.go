package store

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"mokapi/kafka"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	openapi "mokapi/providers/openapi/schema"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
)

type validator struct {
	validators []recordValidator
}

type recordValidator interface {
	Validate(record *kafka.Record) (*KafkaLog, error)
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

func (v *validator) Validate(record *kafka.Record) (l *KafkaLog, err error) {
	if v == nil {
		return &KafkaLog{
			Key:     LogValue{Binary: kafka.Read(record.Key)},
			Message: LogValue{Binary: kafka.Read(record.Value)},
		}, nil
	}

	for _, val := range v.validators {
		l, err = val.Validate(record)
		if err == nil {
			return
		}
	}
	return &KafkaLog{
		Key:     LogValue{Binary: kafka.Read(record.Key)},
		Message: LogValue{Binary: kafka.Read(record.Value)},
	}, err
}

type messageValidator struct {
	messageId string
	msg       *asyncapi3.Message
	key       *schemaValidator
	payload   *schemaValidator
}

func newMessageValidator(messageId string, msg *asyncapi3.Message) *messageValidator {
	v := &messageValidator{messageId: messageId, msg: msg}

	var msgParser encoding.Parser
	switch s := msg.Payload.Value.Schema.(type) {
	case *schema.Schema:
		msgParser = &parser.Parser{Schema: s, ConvertToSortedMap: true}
	case *openapi.Ref:
		mt := media.ParseContentType(msg.ContentType)
		if mt.IsXml() {
			log.Warnf("unsupported payload type: %T", msg.Payload.Value)
		} else {
			msgParser = &parser.Parser{Schema: openapi.ConvertToJsonSchema(s), ConvertToSortedMap: true}
		}
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
		case *schema.Schema:
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

func (mv *messageValidator) Validate(record *kafka.Record) (*KafkaLog, error) {
	r := &KafkaLog{Key: LogValue{}, Message: LogValue{}, Headers: make(map[string]string), MessageId: mv.messageId}

	if mv.payload != nil {
		if v, err := mv.payload.Validate(record.Value); err != nil {
			err = fmt.Errorf("invalid message: %w", err)
			return r, err
		} else {
			b, _ := json.Marshal(v)
			r.Message.Value = string(b)
		}
	} else {
		r.Message.Binary = kafka.Read(record.Value)
	}

	if mv.key != nil {
		if k, err := mv.key.Validate(record.Key); err != nil {
			err = fmt.Errorf("invalid key: %w", err)
			return r, err
		} else {
			b, _ := json.Marshal(k)
			r.Message.Value = string(b)
		}
	} else {
		r.Key.Binary = kafka.Read(record.Key)
	}

	if mv.msg != nil && mv.msg.Bindings.Kafka.SchemaIdLocation != "" {
		switch mv.msg.Bindings.Kafka.SchemaIdLocation {
		case "header":

			// nothing to do
			break
		case "payload":

		default:
			switch mv.msg.ContentType {
			case "avro/binary", "application/octet-stream":
				b := make([]byte, 5)
				_, _ = record.Value.Seek(0, io.SeekStart)
				_, err := record.Value.Read(b)
				if err == nil {
					p := avro.Parser{}
					r.SchemaId, err = p.ParseSchemaId(b)
					if !errors.Is(err, avro.NoSchemaId) {
						return r, err
					}
				}
			default:
				return r, fmt.Errorf("schema id location '%v' not supported", mv.msg.Bindings.Kafka.SchemaIdLocation)
			}
		}
	}

	return r, nil
}

type schemaValidator struct {
	parser      encoding.Parser
	contentType string
}

func (v *schemaValidator) Validate(data kafka.Bytes) (interface{}, error) {
	return encoding.DecodeFrom(data, encoding.WithContentType(media.ParseContentType(v.contentType)), encoding.WithParser(v.parser))
}
