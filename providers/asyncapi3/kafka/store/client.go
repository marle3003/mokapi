package store

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"math/rand"
	"mokapi/config/dynamic"
	"mokapi/kafka"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	openapi "mokapi/providers/openapi/schema"
	"mokapi/runtime/monitor"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/schema"
	"slices"
	"time"

	"github.com/pkg/errors"
)

var TopicNotFound = errors.New("topic not found")
var PartitionNotFound = errors.New("partition not found")

type Record struct {
	Offset         int64          `json:"offset"`
	Key            any            `json:"key"`
	Value          any            `json:"value"`
	Headers        []RecordHeader `json:"headers,omitempty"`
	Partition      int            `json:"partition"`
	SkipValidation bool           `json:"skipValidation,omitempty"`
}

type RecordHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type RecordResult struct {
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
	Headers   []RecordHeader
	Error     string
}

type Client struct {
	ClientId   string
	ScriptFile string

	store   *Store
	monitor *monitor.Kafka
}

func NewClient(s *Store, m *monitor.Kafka) *Client {
	return &Client{
		store:   s,
		monitor: m,
	}
}

func (c *Client) Write(topic string, records []Record, ct media.ContentType) ([]RecordResult, error) {
	t := c.store.Topic(topic)
	if t == nil {
		return nil, TopicNotFound
	}

	var result []RecordResult
	for _, r := range records {
		p, err := c.getPartition(t, r.Partition)
		if err != nil || p == nil {
			return nil, PartitionNotFound
		}
		key, err := c.parseKey(r.Key)
		if err != nil {
			result = append(result, RecordResult{
				Partition: -1,
				Offset:    -1,
				Error:     err.Error(),
			})
		}
		value, err := c.parse(r.Value, ct, p.Topic.Config)
		if err != nil {
			result = append(result, RecordResult{
				Partition: -1,
				Offset:    -1,
				Error:     err.Error(),
			})
			continue
		}
		rec := &kafka.Record{
			Key:   kafka.NewBytes(key),
			Value: kafka.NewBytes(value),
		}
		for _, h := range r.Headers {
			rec.Headers = append(rec.Headers, kafka.RecordHeader{
				Key:   h.Name,
				Value: []byte(h.Value),
			})
		}
		b := kafka.RecordBatch{Records: []*kafka.Record{rec}}

		wr, err := p.write(b, WriteOptions{
			SkipValidation: r.SkipValidation,
			ClientId:       c.ClientId,
			ScriptFile:     c.ScriptFile,
		})
		if err != nil {
			result = append(result, RecordResult{
				Partition: -1,
				Offset:    -1,
				Error:     err.Error(),
			})
		} else {
			if len(wr.Records) > 0 {
				result = append(result, RecordResult{
					Partition: -1,
					Offset:    -1,
					Error:     wr.Records[0].BatchIndexErrorMessage,
				})
			} else {
				rr := RecordResult{
					Offset:    wr.BaseOffset,
					Key:       kafka.Read(b.Records[0].Key),
					Value:     kafka.Read(b.Records[0].Value),
					Partition: p.Index,
				}
				for _, h := range b.Records[0].Headers {
					rr.Headers = append(rr.Headers, RecordHeader{
						Name:  h.Key,
						Value: string(h.Value),
					})
				}
				result = append(result, rr)
				c.store.UpdateMetrics(c.monitor, t, p, b)
			}
		}
	}

	return result, nil
}

func (c *Client) Read(topic string, partition int, offset int64, ct *media.ContentType) ([]Record, error) {
	t := c.store.Topic(topic)
	if t == nil {
		return nil, TopicNotFound
	}
	p := t.Partition(partition)
	if p == nil {
		return nil, PartitionNotFound
	}

	if offset < 0 {
		offset = p.Head
	}

	// read max 6MB
	b, errCode := p.Read(offset, 6e+6)
	if errCode != kafka.None {
		return nil, fmt.Errorf("read records failed: %v", errCode.String())
	}

	records := make([]Record, 0)
	var getValue func(value []byte) (any, error)
	switch {
	case ct.Key() == "application/vnd.mokapi.kafka.binary+json":
		getValue = func(value []byte) (any, error) {
			return base64.StdEncoding.EncodeToString(value), nil
		}
	case ct.Key() == "application/json":
		getValue = func(value []byte) (any, error) {
			var val any
			err := json.Unmarshal(value, &val)
			if err != nil {
				return nil, fmt.Errorf("parse record value as JSON failed: %v", err)
			}
			return val, nil
		}
	default:
		getValue = func(value []byte) (any, error) {
			return string(value), nil
		}
	}

	for _, r := range b.Records {
		key := string(kafka.Read(r.Key))
		val, err := getValue(kafka.Read(r.Value))
		if err != nil {
			return nil, err
		}

		rec := Record{
			Offset:    r.Offset,
			Partition: p.Index,
			Key:       key,
			Value:     val,
		}

		for _, h := range r.Headers {
			rec.Headers = append(rec.Headers, RecordHeader{
				Name:  h.Key,
				Value: string(h.Value),
			})
		}

		records = append(records, rec)
	}

	return records, nil
}

func (c *Client) getPartition(t *Topic, id int) (*Partition, error) {
	if id < 0 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		id = r.Intn(len(t.Partitions))
	} else if id >= len(t.Partitions) {
		return nil, PartitionNotFound
	}

	return t.Partition(id), nil
}

func (c *Client) parse(v any, ct media.ContentType, topic *asyncapi3.Channel) ([]byte, error) {
	if b, ok := v.([]byte); ok {
		return b, nil
	}

	switch ct.Key() {
	case "application/vnd.mokapi.kafka.binary+json":
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("expected string: %v", v)
		}
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("decode base64 string failed: %v", v)
		}
		return b, err
	case "application/vnd.mokapi.kafka.xml+json":
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("expected string: %v", v)
		}
		return []byte(s), nil
	case "application/vnd.mokapi.kafka.json+json":
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("expected string: %v", v)
		}
		return []byte(s), nil
	case "application/json":
		msg, err := selectMessage(v, topic)
		if err != nil {
			return nil, err
		}
		if msg != nil && msg.Payload != nil {
			return msg.Payload.Value.Marshal(v, media.ParseContentType(msg.ContentType))
		}
		b, _ := json.Marshal(v)
		return b, nil
	default:
		msg, err := selectMessage(v, topic)
		if err != nil {
			return nil, err
		}
		if msg != nil && msg.Payload != nil {
			return msg.Payload.Value.Marshal(v, media.ParseContentType(msg.ContentType))
		}

		switch vt := v.(type) {
		case []byte:
			return vt, nil
		default:
			return json.Marshal(v)
		}
	}
}

func (c *Client) parseKey(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	switch vt := v.(type) {
	case []byte:
		return vt, nil
	case string:
		return []byte(vt), nil
	default:
		return json.Marshal(v)
	}
}

func (r *Record) UnmarshalJSON(b []byte) error {
	// set default
	r.Partition = -1

	type alias Record
	a := alias(*r)
	err := dynamic.UnmarshalJSON(b, &a)
	if err != nil {
		return err
	}
	*r = Record(a)
	return nil
}

func selectMessage(value any, topic *asyncapi3.Channel) (*asyncapi3.Message, error) {
	noOperationDefined := true
	var validationErr error
	cfg := topic.Config

	// first try to get send operation
	for _, op := range cfg.Operations {
		if op.Value == nil || op.Value.Channel.Value == nil {
			continue
		}
		if op.Value.Channel.Value == topic && op.Value.Action == "send" {
			noOperationDefined = false
			var messages []*asyncapi3.MessageRef
			if len(op.Value.Messages) == 0 {
				messages = slices.Collect(maps.Values(op.Value.Channel.Value.Messages))
			} else {
				messages = op.Value.Messages
			}
			for _, msg := range messages {
				if msg.Value == nil {
					continue
				}
				if validationErr = valueMatchMessagePayload(value, msg.Value); validationErr == nil {
					return msg.Value, nil
				}
			}
		}
	}

	// second, try to get receive operation
	for _, op := range cfg.Operations {
		if op.Value == nil || op.Value.Channel.Value == nil {
			continue
		}
		if op.Value.Channel.Value == topic && op.Value.Action == "receive" {
			noOperationDefined = false
			var messages []*asyncapi3.MessageRef
			if len(op.Value.Messages) == 0 {
				messages = slices.Collect(maps.Values(op.Value.Channel.Value.Messages))
			} else {
				messages = op.Value.Messages
			}
			for _, msg := range messages {
				if msg.Value == nil {
					continue
				}
				if validationErr = valueMatchMessagePayload(value, msg.Value); validationErr == nil {
					return msg.Value, nil
				}
			}
		}
	}

	if noOperationDefined {
		for _, msg := range topic.Messages {
			if validationErr = valueMatchMessagePayload(value, msg.Value); validationErr == nil {
				return msg.Value, nil
			}
		}
	}

	if value != nil {
		switch value.(type) {
		case string, []byte:
			break
		default:
			b, err := json.Marshal(value)
			if err == nil {
				value = string(b)
			}
		}
		if validationErr != nil {
			return nil, fmt.Errorf("no matching message configuration found for the given value: %v\nhint:\n%w\n", value, validationErr)
		}
		return nil, nil
	}
	return nil, fmt.Errorf("channel defines no message schema; define a message payload in the channel or provide an explicit message")
}

func valueMatchMessagePayload(value any, msg *asyncapi3.Message) error {
	if value == nil || msg.Payload == nil {
		return nil
	}
	ct := media.ParseContentType(msg.ContentType)

	switch v := msg.Payload.Value.Schema.(type) {
	case *schema.Schema:
		_, err := encoding.NewEncoder(v).Write(value, ct)
		return err
	case *openapi.Schema:
		_, err := v.Marshal(value, ct)
		return err
	case *avro.Schema:
		jsSchema := avro.ConvertToJsonSchema(v)
		_, err := encoding.NewEncoder(jsSchema).Write(value, ct)
		return err
	default:
		return nil
	}
}
