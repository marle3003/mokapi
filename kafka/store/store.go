package store

import (
	"fmt"
	"mokapi/kafka/schema"
	"sync"
)

type Store struct {
	brokers map[int]*Broker
	topics  map[string]*Topic
	groups  map[string]*Group

	lock sync.RWMutex
}

type Broker struct {
	id   int
	host string
	port int
}

func New(schema schema.Cluster) *Store {
	c := &Store{
		topics:  make(map[string]*Topic),
		brokers: make(map[int]*Broker),
		groups:  make(map[string]*Group),
	}
	for _, b := range schema.Brokers {
		c.brokers[b.Id] = &Broker{
			id:   b.Id,
			host: b.Host,
			port: b.Port,
		}
	}
	for _, ts := range schema.Topics {
		t, _ := c.addTopic(ts.Name)
		for _, p := range ts.Partitions {
			part := newPartition(p.Index, p.Replicas)
			t.partitions[p.Index] = part

		}
	}
	return c
}

func (s *Store) Topic(name string) *Topic {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if t, ok := s.topics[name]; ok {
		return t
	}
	return nil
}

func (s *Store) NewTopic(name string, numPartitions int) (*Topic, error) {
	t, err := s.addTopic(name)
	if err != nil {
		return t, err
	}
	for i := 0; i < numPartitions; i++ {
		part := newPartition(i, []int{})
		t.partitions[i] = part
	}

	return t, nil
}

func (s *Store) Topics() []*Topic {
	topics := make([]*Topic, 0, len(s.topics))
	for _, t := range s.topics {
		topics = append(topics, t)
	}
	return topics
}

func (s *Store) Broker(id int) (*Broker, bool) {
	b, ok := s.brokers[id]
	return b, ok
}

func (s *Store) Brokers() []*Broker {
	brokers := make([]*Broker, 0, len(s.brokers))
	for _, b := range s.brokers {
		brokers = append(brokers, b)
	}
	return brokers
}

func (s *Store) Groups() []*Group {
	groups := make([]*Group, 0, len(s.groups))
	for _, g := range s.groups {
		groups = append(groups, g)
	}
	return groups
}

func (s *Store) Group(name string) (*Group, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	g, ok := s.groups[name]
	return g, ok
}

func (s *Store) GetOrCreateGroup(name string, brokerId int) *Group {
	s.lock.Lock()
	defer s.lock.Unlock()

	b, ok := s.Broker(brokerId)
	if !ok {
		panic(fmt.Sprintf("unknown broker id: %v", brokerId))
	}

	if g, ok := s.groups[name]; ok {
		return g
	}

	g := &Group{name: name, coordinator: b}
	s.groups[name] = g
	return g
}

func (b *Broker) Id() int {
	return b.id
}

func (b *Broker) Host() string {
	return b.host
}

func (b *Broker) Port() int {
	return b.port
}

func (b *Broker) Addr() string {
	return fmt.Sprintf("%v:%v", b.host, b.port)
}

func (s *Store) addTopic(name string) (*Topic, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.topics[name]; ok {
		return nil, fmt.Errorf("topic %v already exists", name)
	}

	t := &Topic{name: name, partitions: make(map[int]*Partition)}
	s.topics[name] = t

	return t, nil
}
