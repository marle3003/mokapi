package kafka

import (
	"fmt"
	"math/rand"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka"
	"mokapi/config/dynamic/openapi"
	"mokapi/models/media"
	"mokapi/providers/encoding"
	"mokapi/server/kafka/protocol"
	"time"
)

type topic struct {
	name          string
	partitions    map[int]*partition
	payload       *openapi.SchemaRef
	key           *openapi.SchemaRef
	headers       *openapi.SchemaRef
	contentType   string
	config        kafka.TopicBindings
	g             *openapi.Generator
	addedMessage  AddedMessage
	allowedGroups []string
}

func newTopic(name string, c *asyncApi.Channel, leader *broker, addedMessage AddedMessage) *topic {
	msg := c.Publish.Message.Value
	topic := &topic{
		name:         name,
		payload:      msg.Payload,
		partitions:   make(map[int]*partition),
		key:          msg.Bindings.Kafka.Key,
		contentType:  msg.ContentType,
		headers:      msg.Headers,
		g:            openapi.NewGenerator(),
		addedMessage: addedMessage,
		config:       c.Bindings.Kafka,
	}
	if c.Bindings.Kafka.Partitions == 0 {
		topic.partitions[0] = newPartition(0, topic, leader)
	} else {
		for i := 0; i < c.Bindings.Kafka.Partitions; i++ {
			topic.partitions[i] = newPartition(i, topic, leader)
		}
	}

	return topic
}

func (t *topic) addMessage(partition int, key, message interface{}, header interface{}) (interface{}, interface{}, error) {
	record := protocol.Record{Time: time.Now()}

	if key == nil {
		key = t.g.New(t.key)
	}
	record.Key = []byte(fmt.Sprintf("%v", key))

	if message == nil {
		message = t.g.New(t.payload)
	}

	if header == nil {
		header = t.g.New(t.headers)
	}

	var err error
	contentType := media.ParseContentType(t.contentType)
	record.Value, err = encode(message, t.payload, contentType)
	if record.Headers, err = parseHeader(header); err != nil {
		return nil, nil, err
	}

	if err != nil {
		return key, message, err
	}

	if partition < 0 {
		// select random partition
		rand.Seed(time.Now().Unix())
		partition = rand.Intn(len(t.partitions))
	}

	return key, message, t.addRecord(partition, protocol.RecordBatch{
		Records: []protocol.Record{record},
	})
}

func (t *topic) addRecord(partition int, record protocol.RecordBatch) error {
	if partition >= len(t.partitions) {
		return fmt.Errorf("index %q out of range", partition)
	}

	p := t.partitions[partition]

	maxSegmentByes, ok := t.config.SegmentBytes()
	if !ok {
		maxSegmentByes = p.leader.config.LogSegmentBytes()
	}

	if maxSegmentByes >= 0 && p.segments[p.activeSegment].Size > maxSegmentByes {
		p.addNewSegment()
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	record.Offset = p.offset
	p.offset++

	segment := p.segments[p.activeSegment]

	segment.log = append(segment.log, record)
	segment.Size += int(record.Size())
	segment.tail = record.Offset
	segment.lastWritten = time.Now()

	go func() {
		for _, r := range record.Records {
			t.addedMessage(t.name, r.Key, r.Value, partition)
		}
	}()

	return nil
}

func (t *topic) addRecords(partition int, records []protocol.RecordBatch) error {
	for _, r := range records {
		err := t.addRecord(partition, r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *topic) update(config kafka.TopicBindings, leader *broker) {
	t.config = config
	for _, p := range t.partitions {
		p.leader = leader
	}
}

func encode(data interface{}, schema *openapi.SchemaRef, contentType *media.ContentType) ([]byte, error) {
	switch contentType.Subtype {
	case "json":
		return encoding.MarshalJSON(data, schema)
	case "xml", "rss+xml":
		return encoding.MarshalXML(data, schema)
	default:
		if s, ok := data.(string); ok {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("unspupported encoding for content type %v", contentType)
	}
}

func parseHeader(i interface{}) ([]protocol.RecordHeader, error) {
	headers := make([]protocol.RecordHeader, 0)
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type of header: %t", i)
	}

	for k, v := range m {
		if s, ok := v.(string); ok {
			headers = append(headers, protocol.RecordHeader{Key: k, Value: []byte(s)})
		} else {
			return nil, fmt.Errorf("unexpected type of header value %v: %t", k, v)
		}

	}

	return headers, nil
}
