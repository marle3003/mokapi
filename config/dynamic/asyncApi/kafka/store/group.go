package store

import "mokapi/kafka"

type GroupState int

const (
	Empty               GroupState = iota
	PreparingRebalance  GroupState = 1
	CompletingRebalance GroupState = 2
	Stable                         = 3
)

type Group struct {
	Name        string
	Coordinator *Broker
	State       GroupState
	Generation  *Generation

	// todo add timestamp and metadata to commit
	Commits map[string]map[int]int64

	balancer *groupBalancer
}

func NewGroup(name string, coordinator *Broker) *Group {
	g := &Group{
		Name:        name,
		Coordinator: coordinator,
	}
	g.balancer = newGroupBalancer(g, coordinator.kafkaConfig)
	go g.balancer.run()
	return g
}

type Generation struct {
	Id       int
	Protocol string
	LeaderId string
	Members  map[string]*Member

	RebalanceTimeoutMs int
}

type Member struct {
	Partitions     []*Partition
	Client         *kafka.ClientContext
	SessionTimeout int
}

func (g *Group) NewGeneration() *Generation {
	var id int
	if g.Generation == nil {
		id = 0
	} else {
		id = g.Generation.Id + 1
	}
	g.Generation = &Generation{
		Id:      id,
		Members: make(map[string]*Member)}
	return g.Generation

}

func (g *Group) Commit(topic string, partition int, offset int64) {
	if g.Commits == nil {
		g.Commits = make(map[string]map[int]int64)
	}
	t, ok := g.Commits[topic]
	if !ok {
		t = make(map[int]int64)
		g.Commits[topic] = t
	}
	t[partition] = offset
}

func (g *Group) Offset(topic string, partition int) int64 {
	if t, ok := g.Commits[topic]; ok {
		if offset, ok := t[partition]; ok {
			return offset
		}
	}
	// If there is no offset associated with a topic-partition under that consumer group the broker
	// does not set an error code (since it is not really an error), but returns empty metadata and sets the
	// offset field to -1.
	return -1
}

func (g GroupState) String() string {
	switch g {
	case Empty:
		return "Empty"
	case PreparingRebalance:
		return "PreparingRebalance"
	case CompletingRebalance:
		return "CompletingRebalance"
	case Stable:
		return "Stable"
	}
	return "Unknown"
}
