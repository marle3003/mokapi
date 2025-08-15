package store

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"mokapi/config/dynamic"
	"mokapi/kafka"
	"mokapi/media"
	"mokapi/runtime/monitor"
	"time"

	"github.com/pkg/errors"
)

var TopicNotFound = errors.New("topic not found")
var PartitionNotFound = errors.New("partition not found")

type Record struct {
	Key       any
	Value     any
	Headers   map[string]string
	Partition int
}

type RecordResult struct {
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
	Error     string
}

type Client struct {
	store   *Store
	monitor *monitor.Kafka
}

func NewClient(s *Store, m *monitor.Kafka) *Client {
	return &Client{
		store:   s,
		monitor: m,
	}
}

func (c *Client) Write(topic string, records []Record, ct *media.ContentType) ([]RecordResult, error) {
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
		value, err := c.parse(r.Value, ct)
		if err != nil {
			result = append(result, RecordResult{
				Partition: -1,
				Offset:    -1,
				Error:     err.Error(),
			})
		}
		rec := &kafka.Record{
			Key:   kafka.NewBytes(key),
			Value: kafka.NewBytes(value),
		}
		for name, val := range r.Headers {
			rec.Headers = append(rec.Headers, kafka.RecordHeader{
				Key:   name,
				Value: []byte(val),
			})
		}
		b := kafka.RecordBatch{Records: []*kafka.Record{rec}}
		offset, res, err := p.Write(b)
		if err != nil {
			result = append(result, RecordResult{
				Partition: -1,
				Offset:    -1,
				Error:     err.Error(),
			})
		} else {
			if len(res) > 0 {
				result = append(result, RecordResult{
					Partition: -1,
					Offset:    -1,
					Error:     res[0].BatchIndexErrorMessage,
				})
			} else {
				result = append(result, RecordResult{
					Offset:    offset,
					Key:       kafka.Read(b.Records[0].Key),
					Value:     kafka.Read(b.Records[0].Value),
					Partition: p.Index,
				})
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

	records := []Record{}
	switch ct.Key() {
	case "application/vnd.mokapi.kafka.binary+json":
		for _, r := range b.Records {
			var bKey []byte
			base64.StdEncoding.Encode(bKey, kafka.Read(r.Key))
			var bValue []byte
			base64.StdEncoding.Encode(bValue, kafka.Read(r.Value))
			records = append(records, Record{
				Key:   string(bKey),
				Value: string(bValue),
			})
		}
	case "application/json", "":
		for _, r := range b.Records {
			key := string(kafka.Read(r.Key))
			var val any
			err := json.Unmarshal(kafka.Read(r.Value), &val)
			if err != nil {
				return nil, fmt.Errorf("parse record value as JSON failed: %v", err)
			}

			records = append(records, Record{
				Key:   key,
				Value: val,
			})
		}
	default:
		return nil, fmt.Errorf("unknown content type: %v", ct)
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

func (c *Client) parse(v any, ct *media.ContentType) ([]byte, error) {
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
	case "application/json":
		b, _ := json.Marshal(v)
		return b, nil
	default:
		switch vt := v.(type) {
		case []byte:
			return vt, nil
		case string:
			return []byte(vt), nil
		}
	}
	return nil, fmt.Errorf("unknown content type: %v", ct)
}

func (c *Client) parseKey(v any) ([]byte, error) {
	switch vt := v.(type) {
	case []byte:
		return vt, nil
	case string:
		return []byte(vt), nil
	}
	return nil, fmt.Errorf("key not supported: %v", v)
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
