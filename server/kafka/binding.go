package kafka

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/models"
	"mokapi/providers/workflow"
	"time"
)

type AddedMessage func(string, []byte, []byte)

type Binding struct {
	stop         chan bool
	stopCleaner  chan bool
	listen       string
	isRunning    bool
	brokers      []*broker
	groups       map[string]*group
	topics       map[string]*topic
	config       *asyncApi.Config
	kafka        asyncApi.Kafka
	addedMessage AddedMessage
	clients      map[string]*client
}

func NewBinding(c *asyncApi.Config, addedMessage AddedMessage) *Binding {
	s := &Binding{
		stop:         make(chan bool),
		stopCleaner:  make(chan bool),
		groups:       make(map[string]*group),
		topics:       make(map[string]*topic),
		config:       c,
		addedMessage: addedMessage,
		clients:      make(map[string]*client),
	}

	brokerId := 0
	for name, server := range c.Servers {
		b := newBroker(name, brokerId, server.GetHost(), server.GetPort()) // id is 1 based
		s.brokers = append(s.brokers, b)
		s.kafka = server.Bindings.Kafka
		brokerId++
	}

	workflow.RegisterAction("kafka-producer", newProducer(func(topic string, partition int, key, message interface{}) (interface{}, interface{}, error) {
		if t, ok := s.topics[topic]; !ok {
			return key, message, fmt.Errorf("topic %q not found", topic)
		} else {
			return t.addMessage(partition, key, message)
		}
	}))

	return s
}

func (s *Binding) Apply(data interface{}) error {
	config, ok := data.(*asyncApi.Config)
	if !ok {
		return errors.Errorf("unexpected parameter type %T in kafka binding", data)
	}
	s.config = config

	for n, c := range config.Channels {
		name := n[1:] // remove leading slash from name
		if _, ok := s.topics[name]; !ok {
			log.Infof("kafka: adding topic %q", name)
			msg := c.Value.Publish.Message.Value
			s.topics[name] = newTopic(name, c.Value.Bindings.Kafka.Partitions, s.brokers[0], msg.Payload, msg.Bindings.Kafka.Key, msg.ContentType, s.kafka.Log)
		}
	}

	shouldRestart := false
	if s.isRunning {
		log.Infof("Updated configuration of kafka server: %v", s.listen)

		if shouldRestart {
			go s.Start()
		}
	}
	return nil
}

func (s *Binding) Stop() {
	s.stop <- true
	s.stopCleaner <- true
}

func (s *Binding) Start() {
	s.isRunning = true

	for _, b := range s.brokers {
		b.start(s.handle)
	}
}

func (s *Binding) UpdateMetrics(m *models.KafkaMetric) {
	for _, topic := range s.topics {
		var t models.KafkaTopic
		if o, ok := m.Topics[topic.name]; !ok {
			t = models.KafkaTopic{
				Name:       topic.name,
				Partitions: len(topic.partitions),
				Segments:   0,
				Count:      0,
				Size:       0,
			}
			m.Topics[t.Name] = t
		} else {
			t = o
		}

		t.Segments = 0
		t.Count = 0
		t.Size = 0

		for _, p := range topic.partitions {
			t.Segments += len(p.segments)
			t.Count += p.offset
			for _, seg := range p.segments {
				t.Count += seg.tail - seg.head
				t.Size += seg.Size
				if seg.lastWritten.After(t.LastRecord) {
					t.LastRecord = seg.lastWritten
				}
			}
		}

		m.Topics[t.Name] = t
	}
}

func (s *Binding) startCleaner() {
	retentionTime := time.Duration(0)
	if s.kafka.Log.Retention.Hours > 0 {
		retentionTime = time.Duration(s.kafka.Log.Retention.Hours) * time.Hour
	} else if s.kafka.Log.Retention.Minutes > 0 {
		retentionTime = time.Duration(s.kafka.Log.Retention.Minutes) * time.Minute
	} else if s.kafka.Log.Retention.Ms > 0 {
		retentionTime = time.Duration(s.kafka.Log.Retention.Ms) * time.Millisecond
	} else {
		return // no time limit defined
	}

	go func() {
		ticker := time.NewTicker(time.Duration(s.kafka.Log.CleanerBackoffMs) * time.Millisecond)

		for {
			select {
			case <-s.stopCleaner:
				return
			case <-ticker.C:
				now := time.Now()
				for _, t := range s.topics {

					for i, p := range t.partitions {
						partitionSize := int64(0)
						for k, seg := range p.segments {
							partitionSize += seg.Size
							if now.After(seg.lastWritten.Add(retentionTime)) {
								log.Infof("Deleting segment with base offset [%v,%v] from topic %q", seg.head, seg.tail, t.name)
								p.deleteSegment(k)
							}
						}

						if s.kafka.Log.Retention.Bytes > 0 && partitionSize >= s.kafka.Log.Retention.Bytes {
							log.Infof("Maximum partition size reached. Cleanup partition %v from topic %q", i, t.name)
							p.addNewSegment()
							p.deleteAllInactiveSegments()
						}
					}
				}
			}
		}
	}()
}
