package kafka

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/models"
	"mokapi/providers/pipeline"
	"mokapi/providers/pipeline/lang/types"
	"time"
)

type Binding struct {
	stop        chan bool
	stopCleaner chan bool
	listen      string
	isRunning   bool
	brokers     []*broker
	groups      map[string]*group
	topics      map[string]*topic
	config      *asyncApi.Config
	kafka       asyncApi.Kafka
	scheduler   *pipeline.Scheduler
	clients     map[string]*client
}

func NewBinding(c *asyncApi.Config) *Binding {
	s := &Binding{
		stop:        make(chan bool),
		stopCleaner: make(chan bool),
		groups:      make(map[string]*group),
		topics:      make(map[string]*topic),
		config:      c,
		kafka:       c.Info.Kafka,
		clients:     make(map[string]*client),
		scheduler:   pipeline.NewScheduler(),
	}

	brokerId := 0
	for name, b := range c.Servers {
		b := newBroker(name, brokerId, b.GetHost(), b.GetPort()) // id is 1 based
		s.brokers = append(s.brokers, b)
		brokerId++
	}

	return s
}

func (s *Binding) Apply(data interface{}) error {
	config, ok := data.(*asyncApi.Config)
	if !ok {
		return errors.Errorf("unexpected parameter type %T in kafka binding", data)
	}
	s.config = config
	s.kafka = config.Info.Kafka

	s.scheduler.Stop()
	if config.Info.Mokapi != nil {
		err := s.scheduler.Start(config.Info.Mokapi.Value, pipeline.WithSteps(map[string]types.Step{
			"producer": newProducerStep(s.topics),
		}))
		if err != nil {
			return err
		}
	}

	for n, c := range config.Channels {
		name := n[1:] // remove leading slash from name
		if _, ok := s.topics[name]; !ok {
			log.Infof("kafka: adding topic %q", name)
			var msg *asyncApi.Message
			if c.Publish.Message != nil {
				msg = c.Publish.Message.Value
			}
			s.topics[name] = newTopic(name, s.brokers[0], s.kafka.Log, msg)
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
	s.startCleaner()
}

func (s *Binding) UpdateMetrics(m *models.KafkaMetrics) {
	m.Topics = len(s.topics)
	m.Partitions = 0
	m.Segments = 0
	m.Messages = 0
	for _, t := range s.topics {
		m.Partitions += len(t.partitions)
		var size int64
		for _, p := range t.partitions {
			m.Segments += len(p.segments)
			m.Messages += p.offset
			for _, seg := range p.segments {
				size += seg.Size
			}
		}
		m.TopicSizes[t.name] = size
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
