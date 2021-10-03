package kafka

import (
	"fmt"
	"math/rand"
	"mokapi/config/dynamic/asyncApi/kafka"
	"mokapi/config/dynamic/openapi"
	"mokapi/models/media"
	"mokapi/providers/encoding"
	"mokapi/server/kafka/protocol"
	"time"
)

type topic struct {
	name         string
	partitions   map[int]*partition
	payload      *openapi.SchemaRef
	key          *openapi.SchemaRef
	contentType  string
	config       kafka.TopicBindings
	g            *openapi.Generator
	addedMessage AddedMessage
}

func newTopic(name string, config kafka.TopicBindings, leader *broker, payload *openapi.SchemaRef, key *openapi.SchemaRef, contentType string, addedMessage AddedMessage) *topic {
	p := make(map[int]*partition)
	if config.Partitions == 0 {
		p[0] = newPartition(leader)
	} else {
		for i := 0; i < config.Partitions; i++ {
			p[i] = newPartition(leader)
		}
	}
	return &topic{
		name:         name,
		partitions:   p,
		payload:      payload,
		key:          key,
		contentType:  contentType,
		g:            openapi.NewGenerator(),
		addedMessage: addedMessage,
	}
}

func (t *topic) addMessage(partition int, key, message interface{}) (interface{}, interface{}, error) {
	record := protocol.Record{Time: time.Now()}

	if key == nil {
		key = t.g.New(t.key)
	}
	record.Key = []byte(fmt.Sprintf("%v", key))

	if message == nil {
		message = t.g.New(t.payload)
	}

	var err error
	contentType := media.ParseContentType(t.contentType)
	record.Value, err = encode(message, t.payload, contentType)
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

	p.lock.Lock()
	defer p.lock.Unlock()

	record.Offset = p.offset
	p.offset++

	segment := p.segments[p.activeSegment]

	segment.log = append(segment.log, record)
	segment.Size += int64(record.Size())
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
