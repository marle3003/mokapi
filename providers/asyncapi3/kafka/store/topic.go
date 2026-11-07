package store

import (
	"fmt"
	"math/rand"
	"mokapi/kafka"
	"mokapi/kafka/produce"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/schema/json/schema"
	"slices"
	"sync"
	"time"
)

type Topic struct {
	Name string
	// Partitions is replaced when the config changes, use Partition, NumPartitions or PartitionList
	// once the topic is shared.
	Partitions []*Partition

	logger     LogRecord
	s          *Store
	Config     *asyncapi3.Channel
	operations []*asyncapi3.Operation

	// guards Config and Partitions; lock order: Store.m, Topic.m, Partition.m
	m sync.RWMutex
}

type Operation struct {
	GroupId  *schema.Schema
	ClientId *schema.Schema
}

func (t *Topic) Partition(index int) *Partition {
	if t == nil {
		return nil
	}

	t.m.RLock()
	defer t.m.RUnlock()

	if index >= len(t.Partitions) {
		return nil
	}
	return t.Partitions[index]
}

func (t *Topic) NumPartitions() int {
	t.m.RLock()
	defer t.m.RUnlock()
	return len(t.Partitions)
}

// PartitionList returns the current partitions. Later config changes do not modify the returned slice.
func (t *Topic) PartitionList() []*Partition {
	t.m.RLock()
	defer t.m.RUnlock()
	return slices.Clone(t.Partitions)
}

// channelAndPartitions returns Config and Partitions as one consistent pair.
func (t *Topic) channelAndPartitions() (*asyncapi3.Channel, []*Partition) {
	t.m.RLock()
	defer t.m.RUnlock()
	return t.Config, slices.Clone(t.Partitions)
}

func (t *Topic) delete() {
	deletePartitions(t.PartitionList())
}

func newTopic(name string, channel *asyncapi3.Channel, ops []*asyncapi3.Operation, s *Store) *Topic {
	t := &Topic{Name: name, logger: s.log, s: s, Config: channel, operations: ops}

	numPartitions := 1
	if channel.Bindings != nil {
		numPartitions = channel.Bindings.Kafka.Partitions
	}
	for i := 0; i < numPartitions; i++ {
		part := newPartition(i, s.brokers, t.log, s.trigger, t)
		part.validator.Store(newValidator(channel))
		t.Partitions = append(t.Partitions, part)
	}

	return t
}

func (t *Topic) update(config *asyncapi3.Channel) (removed []*Partition) {
	t.m.Lock()
	defer t.m.Unlock()

	t.Config = config
	numPartitions := 1
	if config.Bindings != nil {
		numPartitions = config.Bindings.Kafka.Partitions
	}

	return t.updatePartitionsLocked(numPartitions)
}

// updatePartitions is called outside of Store.Update, so it takes the store lock itself
// (the new partitions are assigned to the store's brokers).
func (t *Topic) updatePartitions(numPartitions int) {
	t.s.m.RLock()
	t.m.Lock()
	removed := t.updatePartitionsLocked(numPartitions)
	t.m.Unlock()
	t.s.m.RUnlock()

	deletePartitions(removed)
}

// updatePartitionsLocked requires Store.m and Topic.m to be held. The caller deletes the returned
// partitions after releasing the locks, so that a script handler running inside a partition
// write can't block the store.
func (t *Topic) updatePartitionsLocked(numPartitions int) (removed []*Partition) {
	n := max(numPartitions, len(t.Partitions))

	for i := range n {
		var p *Partition
		if i >= numPartitions {
			removed = append(removed, t.Partitions[i])
		} else {
			if i >= len(t.Partitions) {
				s := t.Store()
				p = newPartition(i, s.brokers, t.log, s.trigger, t)
				t.Partitions = append(t.Partitions, p)
			} else {
				p = t.Partitions[i]
			}
			p.validator.Store(newValidator(t.Config))
		}
	}

	t.Partitions = t.Partitions[:numPartitions]
	return removed
}

func deletePartitions(partitions []*Partition) {
	for _, p := range partitions {
		p.delete()
	}
}

func (t *Topic) log(r *KafkaMessageLog, traits events.Traits) {
	topicName := t.Config.ResolveAddress()
	traits.With("topic", t.Config.ResolveAddress())
	if t.Name != topicName {
		traits.With("instance", t.Name)
	}
	t.logger(r, traits)
}

func (t *Topic) Store() *Store {
	return t.s
}

func (t *Topic) Write(record *kafka.Record) (partition int, recordError *produce.RecordError) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	index := r.Intn(t.NumPartitions())
	wr, _ := t.Partition(index).Write(kafka.RecordBatch{Records: []*kafka.Record{record}})
	if wr.Records == nil {
		return index, nil
	}
	return index, &wr.Records[0]
}

func (t *Topic) WritePartition(partition int, record *kafka.Record) (recordError *produce.RecordError, err error) {
	p := t.Partition(partition)
	if p == nil {
		return nil, fmt.Errorf("partition out of range")
	}
	wr, _ := p.Write(kafka.RecordBatch{Records: []*kafka.Record{record}})
	if wr.Records == nil {
		return nil, nil
	}
	return &wr.Records[0], nil
}
