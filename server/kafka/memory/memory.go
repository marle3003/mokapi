package memory

import (
	"fmt"
	"mokapi/server/kafka"
	"sync"
)

type Cluster struct {
	brokers map[int]*Broker
	topics  map[string]*Topic
	groups  map[string]*Group

	lock sync.RWMutex
}

type Broker struct {
	id   int
	host string
	port int
}

func (c *Cluster) Topic(name string) kafka.Topic {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if t, ok := c.topics[name]; ok {
		return t
	}
	return nil
}

func (c *Cluster) AddTopic(name string) (kafka.Topic, error) {
	return c.addTopic(name)
}

func (c *Cluster) Topics() []kafka.Topic {
	topics := make([]kafka.Topic, 0, len(c.topics))
	for _, t := range c.topics {
		topics = append(topics, t)
	}
	return topics
}

func (c *Cluster) Brokers() []kafka.Broker {
	brokers := make([]kafka.Broker, 0, len(c.brokers))
	for _, b := range c.brokers {
		brokers = append(brokers, b)
	}
	return brokers
}

func (c *Cluster) Groups() []kafka.Group {
	groups := make([]kafka.Group, 0, len(c.groups))
	for _, g := range c.groups {
		groups = append(groups, g)
	}
	return groups
}

func (c *Cluster) Group(name string) kafka.Group {
	c.lock.Lock()
	defer c.lock.Unlock()

	if g, ok := c.groups[name]; ok {
		return g
	}
	g := &Group{name: name}
	c.groups[name] = g

	return g
}

func (c *Cluster) NewGroup(name string) (kafka.Group, error) {
	return c.newGroup(name)
}

func (b *Broker) Id() int {
	return b.id
}

func (b *Broker) Host() string {
	return b.host
}

func (b *Broker) Port() int {
	return b.port
}

func (c *Cluster) addTopic(name string) (*Topic, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.topics[name]; ok {
		return nil, fmt.Errorf("topic %v already exists", name)
	}

	t := &Topic{name: name, partitions: make(map[int]*Partition)}
	c.topics[name] = t

	return t, nil
}

func (c *Cluster) newGroup(name string) (*Group, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.groups[name]; ok {
		return nil, fmt.Errorf("group %v already exists", name)
	}

	g := &Group{name: name}
	c.groups[name] = g

	return g, nil
}
