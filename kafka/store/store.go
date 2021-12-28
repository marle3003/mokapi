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

func New(config *asyncApi.Config) *Store {
	s := &Store{
		topics:  make(map[string]*Topic),
		brokers: make(map[int]*Broker),
		groups:  make(map[string]*Group),
	}

	for n, server := range config.Servers {
		s.addBroker(n, server)
	}
	for name, ch := range config.Channels {
		if ch.Value == nil {
			continue
		}
		_, _ = s.addTopic(name, ch.Value)
	}
	return s
}

func (s *Store) Topic(name string) *Topic {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if t, ok := s.topics[name]; ok {
		return t
	}
	return nil
}

func (s *Store) NewTopic(name string, config *asyncApi.Channel) (*Topic, error) {
	return s.addTopic(name, config)
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

func (s *Store) Update(c *asyncApi.Config) {
	for n, server := range c.Servers {
		if b := s.getBroker(n); b != nil {
			b.host, b.port = parseHostAndPort(server.Url)
		} else {
			s.addBroker(n, server)
		}
	}
	for _, b := range s.brokers {
		if _, ok := c.Servers[b.name]; !ok {
			s.deleteBroker(b.id)
		}
	}

	for n, ch := range c.Channels {
		k := ch.Value.Bindings.Kafka
		if t, ok := s.topics[n]; ok {
			t.validator.update(ch.Value)
			for _, p := range t.partitions[k.Partitions():] {
				p.delete()
			}
		} else {
			s.addTopic(n, ch.Value)
		}
	}
	for name := range s.topics {
		if _, ok := c.Channels[name]; !ok {
			s.deleteTopic(name)
		}
	}
}

func (s *Store) addTopic(name string, config *asyncApi.Channel) (*Topic, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.topics[name]; ok {
		return nil, fmt.Errorf("topic %v already exists", name)
	}

	k := config.Bindings.Kafka
	t := &Topic{name: name, partitions: make([]*Partition, k.Partitions())}
	s.topics[name] = t

	t.validator = newValidator(config)

	for i := 0; i < k.Partitions(); i++ {
		part := newPartition(i, s.brokers)
		part.validator = t.validator
		t.partitions[i] = part

	}

	return t, nil
}

func (s *Store) deleteTopic(name string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	t, ok := s.topics[name]
	if !ok {
		return
	}
	t.delete()
	delete(s.topics, name)
}

func (s *Store) addBroker(name string, config asyncApi.Server) {
	s.lock.Lock()
	defer s.lock.Unlock()

	id := len(s.brokers)
	b := &Broker{id: id, name: name}
	s.brokers[id] = b
	b.host, b.port = parseHostAndPort(config.Url)
}

func (s *Store) deleteBroker(id int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, t := range s.topics {
		for _, p := range t.partitions {
			p.removeReplica(id)
		}
	}
	delete(s.brokers, id)
}

func (s *Store) getBroker(name string) *Broker {
	for _, b := range s.brokers {
		if b.name == name {
			return b
		}
	}
	return nil
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
