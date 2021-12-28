package store

import (
	"fmt"
	"mokapi/config/dynamic/asyncApi"
	"net"
	"net/url"
	"strconv"
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

func New(config *asyncApi.Config) *Store {
	c := &Store{
		topics:  make(map[string]*Topic),
		brokers: make(map[int]*Broker),
		groups:  make(map[string]*Group),
	}

	brokerId := 0
	replicas := make([]int, 0, len(config.Servers))
	for _, b := range config.Servers {
		host, port := parseHostAndPort(b.Url)
		replicas = append(replicas, brokerId)
		c.brokers[brokerId] = &Broker{
			id:   brokerId,
			host: host,
			port: port,
		}
		brokerId++
	}
	for name, ch := range config.Channels {
		if ch.Value == nil {
			continue
		}

		t, _ := c.addTopic(name)
		t.validator = newValidator(ch.Value)

		k := ch.Value.Bindings.Kafka
		for i := 0; i < k.Partitions(); i++ {
			part := newPartition(i, replicas)
			part.validator = t.validator
			t.partitions[i] = part

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

func parseHostAndPort(s string) (host string, port int) {
	var err error
	var portString string
	host, portString, err = net.SplitHostPort(s)
	if err != nil {
		u, err := url.Parse(s)
		if err != nil || u.Host == "" {
			u, err = url.Parse("//" + s)
			if err != nil {
				return "", 9092
			}
		}

		host = u.Host
		portString = u.Port()
	}

	if len(portString) == 0 {
		port = 9092
	} else {
		var p int64
		p, err = strconv.ParseInt(portString, 10, 32)
		if err != nil {
			return
		}
		port = int(p)
	}

	return
}
