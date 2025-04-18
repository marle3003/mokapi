package store

import (
	"fmt"
	"mokapi/providers/asyncapi3"
	"time"
)

type logCleaner func(broker *Broker)

type Brokers map[int]*Broker

type Broker struct {
	Id   int
	Name string
	Host string
	Port int

	kafkaConfig     asyncapi3.BrokerBindings
	stopCleanerChan chan bool
}

func newBroker(id int, name string, config asyncapi3.Server) *Broker {
	h, p := parseHostAndPort(config.Host)

	return &Broker{
		Id:              id,
		Name:            name,
		Host:            h,
		Port:            p,
		kafkaConfig:     config.Bindings.Kafka,
		stopCleanerChan: make(chan bool, 1),
	}
}

func (b *Broker) Addr() string {
	return fmt.Sprintf("%v:%v", b.Host, b.Port)
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

func (b *Broker) stopCleaner() {
	b.stopCleanerChan <- true
}
