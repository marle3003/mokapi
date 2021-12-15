package memory

import (
	"fmt"
	"mokapi/server/kafka"
)

type Group struct {
	name        string
	coordinator *Broker
	state       kafka.GroupState
	generation  *kafka.Generation

	// todo add timestamp and metadata to commit
	commits map[string]map[int]int64
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Coordinator() (kafka.Broker, error) {
	if g.coordinator == nil {
		return nil, fmt.Errorf("coordinator not set")
	}
	return g.coordinator, nil
}

func (g *Group) State() kafka.GroupState {
	return g.state
}

func (g *Group) SetState(state kafka.GroupState) {
	g.state = state
}

func (g *Group) Generation() *kafka.Generation {
	return g.generation
}

func (g *Group) SetGeneration(generation *kafka.Generation) {
	g.generation = generation
}

func (g *Group) NewGeneration() *kafka.Generation {
	var id int
	if g.generation == nil {
		id = 0
	} else {
		id = g.generation.Id + 1
	}
	g.generation = &kafka.Generation{Id: id, Members: make(map[string]*kafka.Member)}
	return g.generation

}

func (g *Group) Commit(topic string, partition int, offset int64) {
	if g.commits == nil {
		g.commits = make(map[string]map[int]int64)
	}
	t, ok := g.commits[topic]
	if !ok {
		t = make(map[int]int64)
		g.commits[topic] = t
	}
	t[partition] = offset
}

func (g *Group) Offset(topic string, partition int) int64 {
	if t, ok := g.commits[topic]; ok {
		if offset, ok := t[partition]; ok {
			return offset
		}
	}
	return -1
}
