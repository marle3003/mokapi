package store

import (
	"mokapi/config/dynamic/asyncApi"
	kafkaconfig "mokapi/config/dynamic/asyncApi/kafka"
	"mokapi/kafka"
	"mokapi/providers/openapi/schema"
	"mokapi/runtime/events"
)

type Topic struct {
	Name       string
	Partitions []*Partition
	logger     LogRecord
	s          *Store
	config     kafkaconfig.TopicBindings
	Subscribe  Operation
	Publish    Operation
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

func newTopic(name string, config *asyncApi.Channel, brokers Brokers, logger LogRecord, trigger Trigger, s *Store) *Topic {
	t := &Topic{Name: name, logger: logger, s: s, config: config.Bindings.Kafka}

	numPartitions := config.Bindings.Kafka.Partitions
	if numPartitions == 0 {
		numPartitions = 1
	}
	for i := 0; i < numPartitions; i++ {
		part := newPartition(i, brokers, t.log, trigger, t)
		part.validator = newValidator(config)
		t.Partitions = append(t.Partitions, part)
	}

	if config.Subscribe != nil {
		t.Subscribe = Operation{
			GroupId:  config.Subscribe.Bindings.Kafka.GroupId,
			ClientId: config.Subscribe.Bindings.Kafka.ClientId,
		}
	}

	if config.Publish != nil {
		t.Publish = Operation{
			GroupId:  config.Publish.Bindings.Kafka.GroupId,
			ClientId: config.Publish.Bindings.Kafka.ClientId,
		}
	}

	return t
}

func (t *Topic) log(record kafka.Record, partition int, traits events.Traits) {
	t.logger(record, partition, traits.With("topic", t.Name))
}

func (t *Topic) Store() *Store {
	return t.s
}
