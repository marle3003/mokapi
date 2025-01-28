package store

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
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
	"strconv"
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
		if msg.Value == nil {
			continue
		}
		v.validators = append(v.validators, newMessageValidator(id, msg.Value, c))
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
	header    *schemaValidator
}

func newMessageValidator(messageId string, msg *asyncapi3.Message, channel *asyncapi3.Channel) *messageValidator {
	v := &messageValidator{messageId: messageId, msg: msg}

	var msgParser encoding.Parser
	if channel.Bindings.Kafka.ValueSchemaValidation {
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
	}

	if msg.Bindings.Kafka.Key != nil && channel.Bindings.Kafka.KeySchemaValidation {
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

	if msg.Headers != nil {
		var headerParser encoding.Parser
		switch s := msg.Headers.Value.Schema.(type) {
		case *schema.Schema:
			headerParser = &parser.Parser{Schema: s}
		case *avro.Schema:
			headerParser = &avro.Parser{Schema: s}
		default:
			log.Errorf("unsupported header type: %T", msg.Headers.Value)
		}

		if headerParser != nil {
			v.header = &schemaValidator{
				parser: headerParser,
			}
		}
	}

	return v
}

func (mv *messageValidator) Validate(record *kafka.Record) (*KafkaLog, error) {
	r := &KafkaLog{Key: LogValue{}, Message: LogValue{}, Headers: make(map[string]string), MessageId: mv.messageId}

	if mv.msg != nil && mv.msg.Bindings.Kafka.SchemaIdLocation == "payload" {
		var err error
		r.SchemaId, err = readSchemaId(record.Value, mv.msg.Bindings.Kafka.SchemaIdPayloadEncoding)
		if err != nil {
			return nil, err
		}
	}

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

	return r, nil
}

type schemaValidator struct {
	parser      encoding.Parser
	contentType string
}

func (v *schemaValidator) Validate(data io.Reader) (interface{}, error) {
	return encoding.DecodeFrom(data, encoding.WithContentType(media.ParseContentType(v.contentType)), encoding.WithParser(v.parser))
}

func readSchemaId(payload kafka.Bytes, encoding string) (int, error) {
	var schemaIdLen int
	var err error
	switch encoding {
	case "confluent":
		schemaIdLen = 4
	case "":
		schemaIdLen = 4
	default:
		schemaIdLen, err = strconv.Atoi(encoding)
		if err != nil {
			return -1, fmt.Errorf("unsupported schemaIdPayloadEncoding '%v'", encoding)
		}
	}

	// read magic byte
	_, err = payload.Read(make([]byte, 1))
	if err != nil {
		return -1, err
	}

	b := make([]byte, schemaIdLen)
	n, err := payload.Read(b)
	if err != nil {
		return -1, fmt.Errorf("read schemaId failed: %w", err)
	} else if n != schemaIdLen {
		return -1, fmt.Errorf("read schemaId failed; expected %v bytes but read %v bytes", schemaIdLen, n)
	}

	var version int32
	_, err = binary.Decode(b, binary.BigEndian, &version)

	return int(version), err
}
