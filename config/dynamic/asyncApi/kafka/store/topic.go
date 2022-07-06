package store

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/kafka"
	"mokapi/runtime/events"
)

type Topic struct {
	Name       string
	Partitions []*Partition
	logger     LogRecord
	s          *Store
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

func newTopic(name string, config *asyncApi.Channel, brokers Brokers, logger LogRecord, s *Store) *Topic {
	t := &Topic{Name: name, logger: logger, s: s}

	for i := 0; i < config.Bindings.Kafka.Partitions(); i++ {
		part := newPartition(i, brokers, t.log)
		part.validator = newValidator(config)
		t.Partitions = append(t.Partitions, part)
	}
	return t
}

func (t *Topic) log(record kafka.Record, traits events.Traits) {
	t.logger(record, traits.With("topic", t.Name))
}

func (t *Topic) Store() *Store {
	return t.s
}
