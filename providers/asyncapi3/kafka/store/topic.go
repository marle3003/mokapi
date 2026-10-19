package store

import (
	"fmt"
	"math/rand"
	"mokapi/kafka"
	"mokapi/kafka/produce"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/schema/json/schema"
	"time"
)

type Topic struct {
	Name       string
	Partitions []*Partition

	logger     LogRecord
	s          *Store
	Config     *asyncapi3.Channel
	operations []*asyncapi3.Operation
}

type Operation struct {
	GroupId  *schema.Schema
	ClientId *schema.Schema
}

func (t *Topic) Partition(index int) *Partition {
	if index >= len(t.Partitions) {
		return nil
	}
	return t.Partitions[index]
}

func (t *Topic) delete() {
	for _, p := range t.Partitions {
		p.delete()
	}
}

func newTopic(name string, channel *asyncapi3.Channel, ops []*asyncapi3.Operation, s *Store) *Topic {
	t := &Topic{Name: name, logger: s.log, s: s, Config: channel, operations: ops}

	numPartitions := 1
	if channel.Bindings != nil {
		numPartitions = channel.Bindings.Kafka.Partitions
	}
	for i := 0; i < numPartitions; i++ {
		part := newPartition(i, s.brokers, t.log, s.trigger, t)
		part.validator = newValidator(channel)
		t.Partitions = append(t.Partitions, part)
	}

	return t
}

func (t *Topic) update(config *asyncapi3.Channel, s *Store) {
	t.Config = config
	numPartitions := 1
	if config.Bindings != nil {
		numPartitions = config.Bindings.Kafka.Partitions
	}

	t.updatePartitions(numPartitions)
}

func (t *Topic) updatePartitions(numPartitions int) {
	n := numPartitions
	if n < len(t.Partitions) {
		n = len(t.Partitions)
	}

	for i := range n {
		var p *Partition
		if i >= numPartitions {
			p = t.Partitions[i]
			p.delete()
		} else {
			if i >= len(t.Partitions) {
				s := t.Store()
				p = newPartition(i, s.brokers, t.log, s.trigger, t)
				t.Partitions = append(t.Partitions, p)
			} else {
				p = t.Partitions[i]
			}
			p.validator = newValidator(t.Config)
		}
	}

	t.Partitions = t.Partitions[:numPartitions]
}

func (t *Topic) log(r *KafkaMessageLog, traits events.Traits) {
	t.logger(r, traits.With("topic", t.Config.ResolveAddress()))
}

func (t *Topic) Store() *Store {
	return t.s
}

func (t *Topic) Write(record *kafka.Record) (partition int, recordError *produce.RecordError) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	index := r.Intn(len(t.Partitions))
	wr, _ := t.Partitions[index].Write(kafka.RecordBatch{Records: []*kafka.Record{record}})
	if wr.Records == nil {
		return index, nil
	}
	return index, &wr.Records[0]
}

func (t *Topic) WritePartition(partition int, record *kafka.Record) (recordError *produce.RecordError, err error) {
	if partition >= len(t.Partitions) {
		return nil, fmt.Errorf("partition out of range")
	}
	wr, _ := t.Partitions[partition].Write(kafka.RecordBatch{Records: []*kafka.Record{record}})
	if wr.Records == nil {
		return nil, nil
	}
	return &wr.Records[0], nil
}
