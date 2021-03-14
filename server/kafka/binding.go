package kafka

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/providers/pipeline"
	"mokapi/providers/pipeline/lang/types"
	"net"
	"time"
)

type Binding struct {
	stop        chan bool
	stopCleaner chan bool
	listen      string
	isRunning   bool
	brokers     []broker
	groups      map[string]*group
	topics      map[string]*topic
	config      *asyncApi.Config
	binding     asyncApi.KafkaBinding
	scheduler   *pipeline.Scheduler
	clients     map[string]*client
}

func NewBinding(addr string, binding asyncApi.KafkaBinding, c *asyncApi.Config) *Binding {
	s := &Binding{
		stop:        make(chan bool),
		stopCleaner: make(chan bool),
		listen:      addr,
		groups:      make(map[string]*group),
		topics:      make(map[string]*topic),
		config:      c,
		binding:     binding,
		clients:     make(map[string]*client),
		scheduler:   pipeline.NewScheduler(),
	}

	b := newBroker(1, "localhost", 9092) // id is 1 based
	s.brokers = append(s.brokers, b)

	return s
}

func (s *Binding) Apply(data interface{}) error {
	config, ok := data.(*asyncApi.Config)
	if !ok {
		return errors.Errorf("unexpected parameter type %T in kafka binding", data)
	}
	s.config = config

	s.scheduler.Stop()
	if config.Info.Mokapi != nil {
		err := s.scheduler.Start(config.Info.Mokapi.Value, pipeline.WithSteps(map[string]types.Step{
			"producer": &ProducerStep{topics: s.topics},
		}))
		if err != nil {
			return err
		}
	}

	for n := range config.Channels {
		name := n[1:] // remove leading slash from name
		if _, ok := s.topics[name]; !ok {
			log.Infof("kafka: adding topic %q", name)
			s.topics[name] = newTopic(name, s.brokers[0], s.binding.Log)
		}
	}

	shouldRestart := false
	//if s.listen != "" && s.listen != config.Address {
	//	s.stop <- true
	//	shouldRestart = true
	//}
	//
	//s.listen = config.Address
	//s.listen = "0.0.0.0:9092"

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

	l, err := net.Listen("tcp", s.listen)
	if err != nil {
		log.Errorf("Error listening: %v", err.Error())
		return
	}

	log.Infof("Started kafka server on %v", s.listen)

	// Close the listener when the application closes.
	connChannl := make(chan net.Conn)
	close := make(chan bool)
	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-close:
					return
				default:
					log.Errorf("Error accepting: %v", err.Error())
				}
			}
			// Handle connections in a new goroutine.
			connChannl <- conn
		}
	}()

	go func() {
		for {
			select {
			case conn := <-connChannl:
				log.Infof("kafka: new client connected")
				go s.handle(conn)
			case <-s.stop:
				log.Infof("Stopping ldap server on %v", s.listen)
				close <- true
				l.Close()
			}
		}
	}()
	s.startCleaner()
}

func (s *Binding) startCleaner() {
	retentionTime := time.Duration(0)
	if s.binding.Log.Retention.Hours > 0 {
		retentionTime = time.Duration(s.binding.Log.Retention.Hours) * time.Hour
	} else if s.binding.Log.Retention.Minutes > 0 {
		retentionTime = time.Duration(s.binding.Log.Retention.Minutes) * time.Minute
	} else if s.binding.Log.Retention.Ms > 0 {
		retentionTime = time.Duration(s.binding.Log.Retention.Ms) * time.Millisecond
	} else {
		return // no time limit defined
	}

	go func() {
		ticker := time.NewTicker(time.Duration(s.binding.Log.CleanerBackoffMs) * time.Millisecond)

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

						if s.binding.Log.Retention.Bytes > 0 && partitionSize >= s.binding.Log.Retention.Bytes {
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
