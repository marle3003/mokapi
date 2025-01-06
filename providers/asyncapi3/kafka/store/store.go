package store

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/engine/common"
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
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type Store struct {
	brokers      map[int]*Broker
	topics       map[string]*Topic
	groups       map[string]*Group
	cluster      string
	eventEmitter common.EventEmitter

	m sync.RWMutex
}

func NewEmpty(eventEmitter common.EventEmitter) *Store {
	return &Store{
		topics:       make(map[string]*Topic),
		brokers:      make(map[int]*Broker),
		groups:       make(map[string]*Group),
		eventEmitter: eventEmitter,
	}
}

func New(config *asyncapi3.Config, eventEmitter common.EventEmitter) *Store {
	s := NewEmpty(eventEmitter)
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

func (s *Store) NewTopic(name string, config *asyncapi3.Channel, ops []*asyncapi3.Operation) (*Topic, error) {
	return s.addTopic(name, config, ops)
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

func (s *Store) Update(c *asyncapi3.Config) {
	s.cluster = c.Info.Name
	for n, server := range c.Servers {
		if server.Value.Protocol != "" && server.Value.Protocol != "kafka" {
			continue
		}
		if b := s.getBroker(n); b != nil {
			host, port := parseHostAndPort(server.Value.Host)
			if len(host) == 0 {
				log.Errorf("unable to update broker '%v' to cluster '%v': missing host in url '%v'", n, s.cluster, server.Value.Host)
				continue
			}
			b.Host = host
			b.Port = port
		} else {
			s.addBroker(n, *server.Value)
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
		if t, ok := s.topics[n]; ok {
			t.update(ch.Value, s)
		} else {
			if _, err := s.addTopic(n, ch.Value, getOperations(ch.Value, c)); err != nil {
				log.Errorf("unable to add topic '%v' to broker '%v': %v", n, s.cluster, err)
			}
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
		err = fmt.Errorf("kafka: unsupported api key: %v", req.Header.ApiKey)
	}

	if err != nil && err.Error() != "use of closed network connection" {
		panic(fmt.Sprintf("kafka broker: %v", err))
	}
}

func (s *Store) addTopic(name string, channel *asyncapi3.Channel, ops []*asyncapi3.Operation) (*Topic, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.topics[name]; ok {
		return nil, fmt.Errorf("topic %v already exists", name)
	}
	t := newTopic(name, channel, ops, s)
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

func (s *Store) addBroker(name string, config asyncapi3.Server) {
	s.m.Lock()
	defer s.m.Unlock()

	id := len(s.brokers)
	b := newBroker(id, name, config)

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

func (s *Store) log(key, payload interface{}, headers []kafka.RecordHeader, partition int, offset int64, traits events.Traits) {
	events.Push(NewKafkaLog(key, payload, headers, partition, offset), traits.WithNamespace("kafka").WithName(s.cluster))
}

func (s *Store) trigger(record *kafka.Record) {
	h := map[string]string{}
	for _, v := range record.Headers {
		h[v.Key] = string(v.Value)
	}

	r := &EventRecord{
		Offset:  record.Offset,
		Headers: h,
	}

	if record.Key != nil {
		r.Key = kafka.BytesToString(record.Key)
	}
	if record.Value != nil {
		r.Value = kafka.BytesToString(record.Value)
	}

	s.eventEmitter.Emit("kafka", r)

	record.Key = kafka.NewBytes([]byte(r.Key))
	record.Value = kafka.NewBytes([]byte(r.Value))

	if r.Headers == nil {
		record.Headers = nil
	} else {
		// first loop trough array to ensure order of header values
		headers := record.Headers
		for _, h := range headers {
			v, ok := r.Headers[h.Key]
			if ok {
				h.Value = []byte(v)
				delete(r.Headers, h.Key)
			}
		}

		for k, v := range r.Headers {
			record.Headers = append(record.Headers, kafka.RecordHeader{
				Key:   k,
				Value: []byte(v),
			})
		}
	}

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

func getOperations(channel *asyncapi3.Channel, config *asyncapi3.Config) []*asyncapi3.Operation {
	var ops []*asyncapi3.Operation
	for _, op := range config.Operations {
		if op.Value == nil {
			continue
		}
		if op.Value.Channel.Value == channel {
			ops = append(ops, op.Value)
		}
	}
	return ops
}
