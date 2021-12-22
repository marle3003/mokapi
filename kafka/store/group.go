package store

import (
	"fmt"
)

type GroupState int

const (
	Stable       GroupState = iota
	Joining      GroupState = 1
	AwaitingSync GroupState = 2
)

type Group struct {
	name        string
	coordinator *Broker
	state       GroupState
	generation  *Generation

	// todo add timestamp and metadata to commit
	commits map[string]map[int]int64
}

type Generation struct {
	Id       int
	Protocol string
	Members  map[string]*Member
}

type Member struct {
	Partitions []Partition
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Coordinator() (*Broker, error) {
	if g.coordinator == nil {
		return nil, fmt.Errorf("coordinator not set")
	}
	return g.coordinator, nil
}

func (g *Group) State() GroupState {
	return g.state
}

func (g *Group) SetState(state GroupState) {
	g.state = state
}

func (g *Group) Generation() *Generation {
	return g.generation
}

func (g *Group) SetGeneration(generation *Generation) {
	g.generation = generation
}

func (g *Group) NewGeneration() *Generation {
	var id int
	if g.generation == nil {
		id = 0
	} else {
		id = g.generation.Id + 1
	}
	g.generation = &Generation{
		Id:      id,
		Members: make(map[string]*Member)}
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
