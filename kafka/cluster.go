package kafka

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/kafka/protocol"
	"mokapi/kafka/store"
	"sync"
)

type Cluster struct {
	store   *store.Store
	brokers map[int]*Broker

	m sync.RWMutex
}

func NewCluster(cfg *asyncApi.Config) *Cluster {
	c := &Cluster{store: store.New(cfg), brokers: make(map[int]*Broker)}
	return c
}

func (c *Cluster) Start() error {
	if len(c.brokers) > 0 {
		return fmt.Errorf("cluster already started")
	}
	c.m.Lock()
	defer c.m.Unlock()
	for _, bs := range c.store.Brokers() {
		c.addBroker(bs)
	}

	return nil
}

func (c *Cluster) Close() {
	for _, b := range c.brokers {
		b.Close()
	}
}

func (c *Cluster) Update(cfg *asyncApi.Config) {
	c.store.Update(cfg)

	c.m.Lock()
	defer c.m.Unlock()

	for id, b := range c.brokers {
		if _, ok := c.store.Broker(id); !ok {
			b.Close()
			delete(c.brokers, id)
		}
	}
	for _, b := range c.store.Brokers() {
		if _, ok := c.brokers[b.Id()]; !ok {
			c.addBroker(b)
		}
	}
}

func (c *Cluster) addBroker(bs *store.Broker) {
	b := NewBroker(bs.Id(), bs.Addr())
	b.Store = c.store
	c.brokers[bs.Id()] = b
	go func() {
		log.Infof("starting %v", bs.Name())
		err := b.ListenAndServe()
		if err != protocol.ErrServerClosed {
			log.Errorf("unable to start kafka broker on %v: %v", b.Addr, err)
		}
	}()
}
