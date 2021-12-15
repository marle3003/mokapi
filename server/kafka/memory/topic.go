package memory

import (
	"mokapi/server/kafka"
)

type Topic struct {
	name       string
	partitions map[int]*Partition
	leader     *Broker
	replicas   []*Broker
}

func (t *Topic) Name() string {
	return t.name
}

func (t *Topic) Partition(index int) kafka.Partition {
	if p, ok := t.partitions[index]; ok {
		return p
	}
	return nil
}

func (t *Topic) Partitions() []kafka.Partition {
	partitions := make([]kafka.Partition, 0, len(t.partitions))
	for _, p := range t.partitions {
		partitions = append(partitions, p)
	}
	return partitions
}

func (t *Topic) Leader() kafka.Broker {
	return t.leader
}

func (t *Topic) Replicas() []kafka.Broker {
	brokers := make([]kafka.Broker, 0, len(t.replicas))
	for _, b := range t.replicas {
		brokers = append(brokers, b)
	}
	return brokers
}
