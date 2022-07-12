package store

import (
	"fmt"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka"
	"time"
)

type logCleaner func(broker *Broker)

type Brokers map[int]*Broker

type Broker struct {
	Id   int
	Name string
	Host string
	Port int

	kafkaConfig     kafka.BrokerBindings
	stopCleanerChan chan bool
}

func newBroker(id int, name string, config asyncApi.Server) *Broker {
	h, p := parseHostAndPort(config.Url)

	return &Broker{
		Id:          id,
		Name:        name,
		Host:        h,
		Port:        p,
		kafkaConfig: config.Bindings.Kafka,
	}
}

func (b *Broker) Addr() string {
	return fmt.Sprintf("%v:%v", b.Host, b.Port)
}

func (b *Broker) startCleaner(cleaner logCleaner) {
	go func() {
		ticker := time.NewTicker(time.Duration(b.kafkaConfig.LogRetentionCheckIntervalMs()) * time.Millisecond)
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
