package store

import (
	"mokapi/config/dynamic/asyncApi"
	kafkaconfig "mokapi/config/dynamic/asyncApi/kafka"
	"mokapi/kafka"
	"mokapi/runtime/events"
)

type Topic struct {
	Name        string
	Partitions  []*Partition
	logger      LogRecord
	s           *Store
	kafkaConfig kafkaconfig.TopicBindings
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
	t := &Topic{Name: name, logger: logger, s: s, kafkaConfig: config.Bindings.Kafka}

	for i := 0; i < config.Bindings.Kafka.Partitions(); i++ {
		part := newPartition(i, brokers, t.log, trigger, t)
		part.validator = newValidator(config)
		t.Partitions = append(t.Partitions, part)
	}
	return t
}

func (t *Topic) log(record kafka.Record, partition int, traits events.Traits) {
	t.logger(record, partition, traits.With("topic", t.Name))
}

func (t *Topic) Store() *Store {
	return t.s
}
