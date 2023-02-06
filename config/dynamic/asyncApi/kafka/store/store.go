package store

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"mokapi/kafka/createTopics"
	"mokapi/kafka/fetch"
	"mokapi/kafka/findCoordinator"
	"mokapi/kafka/heartbeat"
	"mokapi/kafka/joinGroup"
	"mokapi/kafka/listgroup"
	"mokapi/kafka/metaData"
	"mokapi/kafka/offset"
	"mokapi/kafka/offsetCommit"
	"mokapi/kafka/offsetFetch"
	"mokapi/kafka/produce"
	"mokapi/kafka/syncGroup"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type Store struct {
	brokers map[int]*Broker
	topics  map[string]*Topic
	groups  map[string]*Group
	cluster string

	m sync.RWMutex
}

func NewEmpty() *Store {
	return &Store{
		topics:  make(map[string]*Topic),
		brokers: make(map[int]*Broker),
		groups:  make(map[string]*Group),
	}
}

func New(config *asyncApi.Config) *Store {
	s := NewEmpty()
	s.Update(config)
	return s
}

func (s *Store) Close() {
	for _, g := range s.groups {
		g.balancer.Stop()
	}
}

func (s *Store) Topic(name string) *Topic {
	s.m.RLock()
	defer s.m.RUnlock()

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
	s.m.Lock()
	defer s.m.Unlock()

	g, ok := s.groups[name]
	return g, ok
}

func (s *Store) GetOrCreateGroup(name string, brokerId int) *Group {
	s.m.Lock()
	defer s.m.Unlock()

	b, ok := s.Broker(brokerId)
	if !ok {
		panic(fmt.Sprintf("unknown broker id: %v", brokerId))
	}

	if g, ok := s.groups[name]; ok {
		return g
	}

	g := NewGroup(name, b)
	s.groups[name] = g
	return g
}

func (s *Store) Update(c *asyncApi.Config) {
	s.cluster = c.Info.Name
	for n, server := range c.Servers {
		if b := s.getBroker(n); b != nil {
			host, port := parseHostAndPort(server.Url)
			if len(host) == 0 {
				log.Errorf("unable to update broker '%v' to cluster '%v': missing host in url '%v'", n, s.cluster, server.Url)
				continue
			}
			b.Host = host
			b.Port = port
		} else {
			s.addBroker(n, server)
		}
	}
	for _, b := range s.brokers {
		if _, ok := c.Servers[b.Name]; !ok {
			s.deleteBroker(b.Id)
		}
	}

	for n, ch := range c.Channels {
		if ch.Value == nil {
			continue
		}
		k := ch.Value.Bindings.Kafka
		if t, ok := s.topics[n]; ok {
			for _, p := range t.Partitions[k.Partitions():] {
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

func (s *Store) ServeMessage(rw kafka.ResponseWriter, req *kafka.Request) {
	var err error
	switch req.Message.(type) {
	case *produce.Request:
		err = s.produce(rw, req)
	case *fetch.Request:
		err = s.fetch(rw, req)
	case *offset.Request:
		err = s.offset(rw, req)
	case *metaData.Request:
		err = s.metadata(rw, req)
	case *offsetCommit.Request:
		err = s.offsetCommit(rw, req)
	case *offsetFetch.Request:
		err = s.offsetFetch(rw, req)
	case *findCoordinator.Request:
		err = s.findCoordinator(rw, req)
	case *joinGroup.Request:
		err = s.joingroup(rw, req)
	case *heartbeat.Request:
		err = s.heartbeat(rw, req)
	case *syncGroup.Request:
		err = s.syncgroup(rw, req)
	case *listgroup.Request:
		err = s.listgroup(rw, req)
	case *apiVersion.Request:
		err = s.apiversion(rw, req)
	case *createTopics.Request:
		err = s.createtopics(rw, req)
	default:
		err = fmt.Errorf("unsupported api key: %v", req.Header.ApiKey)
	}

	if err != nil && err.Error() != "use of closed network connection" {
		panic(fmt.Sprintf("kafka broker: %v", err))
	}
}

func (s *Store) addTopic(name string, config *asyncApi.Channel) (*Topic, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.topics[name]; ok {
		return nil, fmt.Errorf("topic %v already exists", name)
	}
	t := newTopic(name, config, s.brokers, s.log, s)
	s.topics[name] = t
	return t, nil
}

func (s *Store) deleteTopic(name string) {
	s.m.Lock()
	defer s.m.Unlock()

	t, ok := s.topics[name]
	if !ok {
		return
	}
	t.delete()
	delete(s.topics, name)
}

func (s *Store) addBroker(name string, config asyncApi.Server) {
	s.m.Lock()
	defer s.m.Unlock()

	id := len(s.brokers)
	b := newBroker(id, name, config)

	if len(b.Host) == 0 {
		log.Errorf("unable to add broker '%v' to cluster '%v': missing host in url '%v'", name, s.cluster, config.Url)
		return
	}

	s.brokers[id] = b
	b.startCleaner(s.cleanLog)
}

func (s *Store) deleteBroker(id int) {
	s.m.Lock()
	defer s.m.Unlock()

	for _, t := range s.topics {
		for _, p := range t.Partitions {
			p.removeReplica(id)
		}
	}
	if b, ok := s.brokers[id]; ok {
		b.stopCleaner()
	}
	delete(s.brokers, id)
}

func (s *Store) getBroker(name string) *Broker {
	for _, b := range s.brokers {
		if b.Name == name {
			return b
		}
	}
	return nil
}

func (s *Store) getBrokerByHost(addr string) *Broker {
	for _, b := range s.brokers {
		_, p := parseHostAndPort(addr)
		if b.Port == p {
			return b
		}
	}
	return nil
}

func (s *Store) log(record kafka.Record, traits events.Traits) {
	events.Push(NewKafkaLog(record), traits.WithNamespace("kafka").WithName(s.cluster))
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

func (s *Store) UpdateMetrics(m *monitor.Kafka, topic *Topic, partition *Partition, batch kafka.RecordBatch) {
	m.Messages.WithLabel(s.cluster, topic.Name).Add(float64(len(batch.Records)))
	m.LastMessage.WithLabel(s.cluster, topic.Name).Set(float64(time.Now().Unix()))

	for name, g := range s.groups {
		gt, ok := g.Commits[topic.Name]
		if !ok {
			continue
		}
		commit, ok := gt[partition.Index]
		if !ok {
			continue
		}
		lag := float64(partition.Offset() - commit)
		m.Lags.WithLabel(s.cluster, name, topic.Name, strconv.Itoa(partition.Index)).Set(lag)
	}
}
