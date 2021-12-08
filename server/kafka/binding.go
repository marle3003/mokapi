package kafka

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/openapi"
	"mokapi/models"
	"mokapi/providers/utils"
	"net"
	"sync"
	"time"
)

type AddedMessage func(topic string, key []byte, message []byte, partition int)

type controller interface {
	handle(conn net.Conn)
	checkRetention(b *broker)
}

type Binding struct {
	listen    string
	isRunning bool
	brokers   map[string]*broker
	groups    map[string]*group
	topics    map[string]*topic
	Config    *asyncApi.Config
	//kafka        kafka.BrokerBindings
	addedMessage AddedMessage
	clients      map[string]*client

	clientsMutex sync.RWMutex
	groupsMutex  sync.RWMutex
	brokerMutex  sync.RWMutex

	nextBrokerId int
}

func NewBinding(addedMessage AddedMessage) *Binding {
	s := &Binding{
		brokers:      make(map[string]*broker),
		groups:       make(map[string]*group),
		topics:       make(map[string]*topic),
		addedMessage: addedMessage,
		clients:      make(map[string]*client),
	}

	return s
}

func (s *Binding) AddMessage(topic string, partition int, key, message, header interface{}) (interface{}, interface{}, error) {
	if t, ok := s.topics[topic]; !ok {
		return key, message, fmt.Errorf("topic %q not found", topic)
	} else {
		return t.addMessage(partition, key, message, header)
	}
}

func (s *Binding) Apply(data interface{}) error {
	config, ok := data.(*asyncApi.Config)
	if !ok {
		return fmt.Errorf("unexpected parameter type %T in kafka binding", data)
	}
	s.Config = config

	s.updateBrokers(config.Servers)
	leader := s.selectLeader()

	for name, c := range config.Channels {
		if topic, ok := s.topics[name]; !ok {
			if c.Value == nil || c.Value.Publish == nil {
				log.Errorf("channel %v: subscribe and publish are required", name)
				continue
			}
			if c.Value.Publish.Message == nil || c.Value.Publish.Message.Value == nil {
				log.Errorf("kafka: message reference error for channel %v", name)
				continue
			}
			if err := validateTopicName(name); err == nil {
				s.topics[name] = newTopic(name, c.Value, leader, s.addedMessage)
			} else {
				log.Errorf("unable to add topic %q: %v", name, err)
			}
			log.Infof("kafka: added topic %q with %v partitions on broker %v:%v", name, c.Value.Bindings.Kafka.Partitions(), leader.host, leader.port)
		} else {
			topic.update(c.Value.Bindings.Kafka, leader)
		}
		if c.Value.Subscribe.Bindings.Kafka.GroupId != nil {
			groupId := c.Value.Subscribe.Bindings.Kafka.GroupId
			for _, g := range getGroupIds(groupId) {
				s.getOrCreateGroup(g)
				for _, p := range s.topics[name].partitions {
					p.committed[g] = 0
				}
			}
		}
	}

	for _, g := range s.groups {
		g.coordinator = leader
	}

	return nil
}

func (s *Binding) Stop() {
	s.brokerMutex.RLock()
	defer s.brokerMutex.RUnlock()

	for _, b := range s.brokers {
		b.stop()
	}
}

func (s *Binding) Start() {
}

func (s *Binding) updateBrokers(servers map[string]asyncApi.Server) {
	s.brokerMutex.Lock()
	defer s.brokerMutex.Unlock()

	for name, broker := range s.brokers {
		if server, ok := servers[name]; !ok {
			broker.stop()
			delete(s.brokers, name)
		} else {
			host, port := server.GetHost(), server.GetPort()
			if broker.host != host || broker.port != port {
				broker.stop()
				broker.host, broker.port = host, port
				broker.start(s)
			}
			broker.config = server.Bindings.Kafka
		}
	}

	for name, server := range servers {
		if _, ok := s.brokers[name]; ok {
			continue
		}
		b := newBroker(s.nextBrokerId, name, server.GetHost(), server.GetPort(), server.Bindings.Kafka)
		s.brokers[name] = b
		b.start(s)
		s.nextBrokerId++
	}
}

func (s *Binding) selectLeader() *broker {
	s.brokerMutex.RLock()
	defer s.brokerMutex.RUnlock()

	for _, broker := range s.brokers {
		return broker
	}
	return nil
}

func (s *Binding) HasBroker(address string) bool {
	s.brokerMutex.RLock()
	defer s.brokerMutex.RUnlock()

	host, port, err := utils.ParseUrl(address)
	if err != nil {
		return false
	}

	for _, broker := range s.brokers {
		if broker.host == host && broker.port == port {
			return true
		}
	}
	return false
}

func (s *Binding) UpdateMetrics(m *models.KafkaMetric) {
	for _, topic := range s.topics {
		var t *models.KafkaTopic
		if o, ok := m.Topics[topic.name]; !ok {
			t = &models.KafkaTopic{
				Service:    s.Config.Info.Name,
				Name:       topic.name,
				Partitions: make(map[int]*models.KafkaPartition),
				Count:      0,
				Groups:     make(map[string]*models.KafkaTopicGroup),
			}
			m.Topics[t.Name] = t
		} else {
			t = o
			t.Service = s.Config.Info.Name
		}

		for name, group := range t.Groups {
			group.Lag = 0
			if g, ok := s.groups[name]; ok {
				g.updateMetrics(&group.KafkaGroup)
			}
		}

		for _, partition := range topic.partitions {
			partition.lock.RLock()

			p, ok := t.Partitions[partition.index]
			if !ok {
				p = &models.KafkaPartition{Index: partition.index}
				t.Partitions[partition.index] = p
			}

			p.Segments = len(partition.segments)
			p.StartOffset = partition.startOffset
			p.Offset = partition.offset
			p.Leader = partition.leader.name
			p.Size = 0

			for name, g := range t.Groups {
				if _, ok := partition.committed[name]; !ok {
					g.Lag += partition.offset
				}
			}
			for name, committed := range partition.committed {
				i, ok := t.Groups[name]
				if !ok {
					i = &models.KafkaTopicGroup{}
					t.Groups[name] = i
				}
				i.Lag += partition.offset - committed
			}

			for _, seg := range partition.segments {
				p.Size += int64(seg.Size)
				if seg.lastWritten.After(t.LastRecord) {
					t.LastRecord = seg.lastWritten
				}
			}

			partition.lock.RUnlock()
		}

		m.Topics[t.Name] = t
	}

	s.groupsMutex.RLock()
	for _, g := range s.groups {
		group, ok := m.Groups[g.name]
		if !ok {
			group = &models.KafkaGroup{Members: make([]string, 0)}
			m.Groups[g.name] = group
		}
		g.updateMetrics(group)
		m.Groups[g.name] = group
	}
	s.groupsMutex.RUnlock()
}

func (s *Binding) checkRetention(b *broker) {
	brokerRetentionTime := time.Duration(b.config.LogRetentionMs()) * time.Millisecond
	brokerRetentionBytes := b.config.LogRetentionBytes()
	brokerRollingTime := time.Duration(b.config.LogRollMs()) * time.Millisecond
	now := time.Now()

	for _, t := range s.topics {
		retentionTime := brokerRetentionTime
		retentionBytes := brokerRetentionBytes
		rollingTime := brokerRollingTime

		if ms, ok := t.config.RetentionMs(); ok {
			retentionTime = time.Duration(ms) * time.Millisecond
		}
		if bytes, ok := t.config.RetentionBytes(); ok {
			retentionBytes = bytes
		}
		if ms, ok := t.config.SegmentMs(); ok {
			rollingTime = time.Duration(ms) * time.Millisecond
		}

		for i, p := range t.partitions {
			if p.leader != b {
				continue
			}

			partitionSize := int64(0)
			for k, seg := range p.segments {
				partitionSize += int64(seg.Size)

				// check rolling
				if now.After(seg.opened.Add(rollingTime)) {
					p.addNewSegment(now)
				}

				// check retention
				if seg.Size > 0 && !seg.closed.IsZero() && now.After(seg.closed.Add(retentionTime)) {
					log.Infof("kafka: deleting segment with base offset [%v,%v] from partition %v topic %q", seg.head, seg.tail, p.index, t.name)
					p.deleteSegment(k)
				}
			}

			if retentionBytes > 0 && partitionSize >= retentionBytes {
				log.Infof("kafka: maximum partition size reached. cleanup partition %v from topic %q", i, t.name)
				p.deleteClosedSegments()
			}
		}
	}
}

func (s *Binding) getGroup(name string) (*group, bool) {
	s.groupsMutex.RLock()
	defer s.groupsMutex.RUnlock()

	g, ok := s.groups[name]
	return g, ok
}

func (s *Binding) getOrCreateGroup(name string) *group {
	s.groupsMutex.Lock()
	defer s.groupsMutex.Unlock()

	g, ok := s.groups[name]
	if !ok {
		b := s.selectLeader()
		g = b.newGroup(name)
		s.groups[name] = g
	}
	return g
}

func getGroupIds(schema *openapi.Schema) (groupIds []string) {
	if schema.Enum != nil {
		for _, i := range schema.Enum {
			if s, ok := i.(string); ok {
				groupIds = append(groupIds, s)
			}
		}
	}
	return
}
