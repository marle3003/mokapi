package store

import (
	"fmt"
	"mokapi/providers/asyncapi3"
	"sync/atomic"
	"time"
)

type logCleaner func(broker *Broker)

type Brokers []*Broker

type hostPort struct {
	host string
	port int
}

type Broker struct {
	Id   int
	Name string

	// host and port change together on config reload, so they are read as one pair.
	addr atomic.Pointer[hostPort]

	config          *asyncapi3.Server
	kafkaConfig     asyncapi3.BrokerBindings
	stopCleanerChan chan bool
	topics          map[string]*Topic
}

func newBroker(id int, name string, config *asyncapi3.Server) *Broker {
	h, p := parseHostAndPort(config.Host)

	bindings := &asyncapi3.ServerBindings{}
	if config.Bindings != nil {
		bindings = config.Bindings
	}

	b := &Broker{
		Id:              id,
		Name:            name,
		config:          config,
		kafkaConfig:     bindings.Kafka,
		stopCleanerChan: make(chan bool, 1),
	}
	b.setHostPort(h, p)
	return b
}

// HostPort returns host and port as one consistent pair, empty and 0 if none is set.
func (b *Broker) HostPort() (string, int) {
	a := b.addr.Load()
	if a == nil {
		return "", 0
	}
	return a.host, a.port
}

func (b *Broker) setHostPort(host string, port int) {
	b.addr.Store(&hostPort{host: host, port: port})
}

func (b *Broker) Addr() string {
	host, port := b.HostPort()
	return fmt.Sprintf("%v:%v", host, port)
}

func (b *Broker) startCleaner(cleaner logCleaner) {
	go func() {
		ms := b.kafkaConfig.LogRetentionCheckIntervalMs
		if ms == 0 {
			ms = 300000 // 5 minutes
		}
		ticker := time.NewTicker(time.Duration(ms) * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-b.stopCleanerChan:
				return
			case <-ticker.C:
				cleaner(b)
			}
		}
	}()
}

// stopCleaner does not block if the cleaner was already told to stop.
func (b *Broker) stopCleaner() {
	select {
	case b.stopCleanerChan <- true:
	default:
	}
}
