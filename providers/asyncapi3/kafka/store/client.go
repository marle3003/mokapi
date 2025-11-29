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
		for _, h := range r.Headers {
			rec.Headers = append(rec.Headers, kafka.RecordHeader{
				Key:   h.Name,
				Value: []byte(h.Value),
			})
		}
		b := kafka.RecordBatch{Records: []*kafka.Record{rec}}
		var write func(batch kafka.RecordBatch) (WriteResult, error)
		if r.SkipValidation {
			write = p.WriteSkipValidation
		} else {
			write = p.Write
		}

		wr, err := write(b)
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

	var records []Record
	var getValue func(value []byte) (any, error)
	switch {
	case ct.Key() == "application/vnd.mokapi.kafka.binary+json":
		getValue = func(value []byte) (any, error) {
			return base64.StdEncoding.EncodeToString(value), nil
		}
	case ct.Key() == "application/json", ct.IsAny():
		getValue = func(value []byte) (any, error) {
			var val any
			err := json.Unmarshal(value, &val)
			if err != nil {
				return nil, fmt.Errorf("parse record value as JSON failed: %v", err)
			}
			return val, nil
		}

	default:
		return nil, fmt.Errorf("unknown content type: %v", ct)
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

func (c *Client) parse(v any, ct media.ContentType) ([]byte, error) {
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
		b, ok := v.([]byte)
		if ok {
			return b, nil
		}
		b, _ = json.Marshal(v)
		return b, nil
	default:
		switch vt := v.(type) {
		case []byte:
			return vt, nil
		case string:
			return []byte(vt), nil
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
