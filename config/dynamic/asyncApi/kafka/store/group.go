package store

import "mokapi/kafka"

type GroupState int

const (
	Stable       GroupState = iota
	Joining      GroupState = 1
	AwaitingSync GroupState = 2
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
	g.balancer = newGroupBalancer(g)
	go g.balancer.run()
	return g
}

type Generation struct {
	Id       int
	Protocol string
	LeaderId string
	Members  map[string]*Member
}

type Member struct {
	Partitions []*Partition
	Client     *kafka.ClientContext
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
	case Joining:
		return "Joining"
	case Stable:
		return "Stable"
	case AwaitingSync:
		return "AwaitingSync"
	}
	return "Unknown"
}
