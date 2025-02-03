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
	"slices"
	"strconv"
	"unicode/utf8"
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
			Headers: convertHeader(record.Headers),
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
		Headers: convertHeader(record.Headers),
	}, err
}

type messageValidator struct {
	messageId string
	msg       *asyncapi3.Message
	key       *schemaValidator
	payload   *schemaValidator
	header    *headerValidator
}

func newMessageValidator(messageId string, msg *asyncapi3.Message, channel *asyncapi3.Channel) *messageValidator {
	v := &messageValidator{messageId: messageId, msg: msg}

	var msgParser encoding.Parser
	if msg.Payload != nil && channel.Bindings.Kafka.ValueSchemaValidation {
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
		case *asyncapi3.AvroRef:
			msgParser = &avro.Parser{Schema: s.Schema}
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
		case *asyncapi3.AvroRef:
			keyParser = &avro.Parser{Schema: s.Schema}
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
		case *asyncapi3.AvroRef:
			headerParser = &avro.Parser{Schema: s.Schema}
		default:
			log.Errorf("unsupported header type: %T", msg.Headers.Value)
		}

		if headerParser != nil {
			v.header = &headerValidator{
				schema: msg.Headers,
			}
		}
	} else {
		v.header = &headerValidator{}
	}

	return v
}

func (mv *messageValidator) Validate(record *kafka.Record) (*KafkaLog, error) {
	r := &KafkaLog{Key: LogValue{}, Message: LogValue{}, Headers: make(map[string]LogValue), MessageId: mv.messageId}

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

	if mv.header != nil {
		if h, err := mv.header.Validate(record.Headers); err != nil {
			err = fmt.Errorf("invalid key: %w", err)
			return r, err
		} else {
			r.Headers = h
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

type headerValidator struct {
	schema *asyncapi3.SchemaRef
}

func (v *headerValidator) Validate(headers []kafka.RecordHeader) (map[string]LogValue, error) {
	return parseHeader(headers, v.schema)
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

func convertHeader(headers []kafka.RecordHeader) map[string]LogValue {
	result := map[string]LogValue{}
	for _, header := range headers {
		result[header.Key] = LogValue{Binary: header.Value}
	}
	return result
}

func parseHeader(headers []kafka.RecordHeader, sr *asyncapi3.SchemaRef) (map[string]LogValue, error) {
	if sr == nil || sr.Value == nil || sr.Value.Schema == nil {
		return convertHeader(headers), nil
	}
	m := map[string][]byte{}
	for _, header := range headers {
		m[header.Key] = header.Value
	}

	result := map[string]LogValue{}
	switch s := sr.Value.Schema.(type) {
	case *schema.Schema:
		if s.Properties == nil {
			return result, fmt.Errorf("invalid header definition: expected object with properties")
		}
		done := map[string]bool{}
		p := parser.Parser{}
		for it := s.Properties.Iter(); it.Next(); {
			if b, ok := m[it.Key()]; ok {
				t := it.Value().Type
				var v interface{}
				var err error
				switch {
				case t.IsString():
					v = string(b)
				case t.IsInteger():
					switch len(b) {
					case 1:
						var i int8
						_, err = binary.Decode(b, binary.LittleEndian, &i)
						v = int(i)
					case 2:
						var i int16
						_, err = binary.Decode(b, binary.LittleEndian, &i)
						v = int(i)
					case 4:
						var i int32
						_, err = binary.Decode(b, binary.LittleEndian, &i)
						v = int(i)
					case 8:
						var i int64
						_, err = binary.Decode(b, binary.LittleEndian, &i)
						v = int(i)
					default:
						return nil, fmt.Errorf("invalid header %v: expected integer got: %v", it.Key(), b)
					}
				case t.IsNumber():
					var f float32
					_, err = binary.Decode(b, binary.LittleEndian, &f)
					v = float64(f)
				}
				if err != nil {
					return nil, err
				}
				v, err = p.ParseWith(v, it.Value())
				if err != nil {
					return nil, err
				}
				result[it.Key()] = LogValue{Value: fmt.Sprintf("%v", v)}
				done[it.Key()] = true
			} else if slices.Contains(s.Required, it.Key()) {
				return nil, fmt.Errorf("required property '%s' is missing in header", it.Key())
			}
		}
		if len(done) != len(headers) && !s.IsFreeForm() {
			var additional []string
			for _, h := range headers {
				if _, ok := done[h.Key]; !ok {
					additional = append(additional, h.Key)
				}
			}
			return nil, fmt.Errorf("additional headers not allowed: %v", additional)
		}
		for _, h := range headers {
			if _, ok := done[h.Key]; !ok {
				if utf8.Valid(h.Value) {
					result[h.Key] = LogValue{Value: string(h.Value), Binary: h.Value}
				} else {
					result[h.Key] = LogValue{Binary: h.Value}
				}
			}
		}
	case *asyncapi3.AvroRef:
		for _, f := range s.Fields {
			if v, ok := m[f.Name]; ok {
				val, err := encoding.Decode(v, encoding.WithParser(&avro.Parser{Schema: s.Schema}))
				if err != nil {
					return nil, err
				}
				b, _ := json.Marshal(val)
				result[f.Name] = LogValue{Value: string(b)}
			}
		}
	}
	return result, nil
}
