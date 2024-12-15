package store

import (
	"mokapi/kafka"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/schema/json/schema"
)

type Topic struct {
	Name       string
	Partitions []*Partition

	logger     LogRecord
	s          *Store
	channel    *asyncapi3.Channel
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

func newTopic(name string, channel *asyncapi3.Channel, ops []*asyncapi3.Operation, brokers Brokers, logger LogRecord, trigger Trigger, s *Store) *Topic {
	t := &Topic{Name: name, logger: logger, s: s, channel: channel, operations: ops}

	numPartitions := channel.Bindings.Kafka.Partitions
	if numPartitions == 0 {
		numPartitions = 1
	}
	for i := 0; i < numPartitions; i++ {
		part := newPartition(i, brokers, t.log, trigger, t)
		part.validator = newValidator(channel)
		t.Partitions = append(t.Partitions, part)
	}

	return t
}

func (t *Topic) log(record *kafka.Record, partition int, traits events.Traits) {
	t.logger(record, partition, traits.With("topic", t.Name))
}

func (t *Topic) Store() *Store {
	return t.s
}
