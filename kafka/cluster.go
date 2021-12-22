package kafka

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka/schema"
	"mokapi/kafka/store"
)

type Cluster struct {
	store   *store.Store
	brokers map[int]*Broker
}

func NewCluster(schema schema.Cluster) *Cluster {
	c := &Cluster{store: store.New(schema), brokers: make(map[int]*Broker)}
	return c
}

func (c *Cluster) Start() error {
	if len(c.brokers) > 0 {
		return fmt.Errorf("cluster already started")
	}
	for _, bs := range c.store.Brokers() {
		b := NewBroker(bs.Id(), bs.Addr())
		b.Store = c.store
		c.brokers[bs.Id()] = b
		log.Infof("starting kafka broker %v", bs.Addr())
		go b.ListenAndServe()
	}

	return nil
}

func (c *Cluster) Close() {
	for _, b := range c.brokers {
		b.Close()
	}
}

func (c *Cluster) Update(schema schema.Cluster) {
	// todo implement
	panic("feature not implemented")
}
