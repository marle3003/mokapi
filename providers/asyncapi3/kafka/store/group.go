package store

import (
	"maps"
	"mokapi/kafka"
	"mokapi/runtime/monitor"
	"slices"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type GroupState int

const (
	Empty               GroupState = iota
	PreparingRebalance  GroupState = 1
	CompletingRebalance GroupState = 2
	Stable                         = 3
)

var states = [...]string{
	Empty:               "Empty",
	PreparingRebalance:  "PreparingRebalance",
	CompletingRebalance: "CompletingRebalance",
	Stable:              "Stable",
}

// Group is shared between the balancer goroutine and request handlers.
// All mutable state is behind m; use the accessor methods.
type Group struct {
	Name string

	m          sync.RWMutex
	state      GroupState
	generation *Generation

	// todo add timestamp and metadata to commit
	commits map[string]map[int]int64

	balancer *groupBalancer
}

// GroupSnapshot is an independent copy of a group's state at one point in time.
// Member.Client is shared with the live group.
type GroupSnapshot struct {
	Name       string
	State      GroupState
	Generation *Generation
	Commits    map[string]map[int]int64
}

func (s *Store) newGroup(name string, coordinator *Broker) *Group {
	g := &Group{
		Name: name,
	}
	g.balancer = newGroupBalancer(g, coordinator.kafkaConfig, &groupMonitor{cluster: s.config.Info.Name, monitor: s.monitor})
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
	Partitions     map[string][]int
	Client         *kafka.ClientContext
	SessionTimeout int
}

func (g *Group) State() GroupState {
	g.m.RLock()
	defer g.m.RUnlock()
	return g.state
}

func (g *Group) SetState(state GroupState) {
	g.m.Lock()
	defer g.m.Unlock()
	g.state = state
}

// Generation returns the current generation or nil. The generation is published once
// and not modified afterwards, except member partitions (see AssignPartitions).
func (g *Group) Generation() *Generation {
	g.m.RLock()
	defer g.m.RUnlock()
	return g.generation
}

func (g *Group) SetGeneration(generation *Generation) {
	g.m.Lock()
	defer g.m.Unlock()
	g.generation = generation
}

// NewGeneration returns the successor of the current generation without publishing it.
// The caller fills it and publishes it with SetGeneration.
func (g *Group) NewGeneration() *Generation {
	g.m.RLock()
	defer g.m.RUnlock()

	var id int
	if g.generation != nil {
		id = g.generation.Id + 1
	}
	return &Generation{
		Id:      id,
		Members: make(map[string]*Member)}
}

// AssignPartitions stores the partitions the leader assigned to a member of the current generation.
func (g *Group) AssignPartitions(memberName string, topics map[string][]int) {
	g.m.Lock()
	defer g.m.Unlock()

	member := g.generation.Members[memberName]
	for topicName, partitions := range topics {
		member.Partitions[topicName] = partitions
	}
}

func (g *Group) Snapshot() GroupSnapshot {
	g.m.RLock()
	defer g.m.RUnlock()

	snapshot := GroupSnapshot{
		Name:    g.Name,
		State:   g.state,
		Commits: make(map[string]map[int]int64, len(g.commits)),
	}
	for topic, partitions := range g.commits {
		snapshot.Commits[topic] = maps.Clone(partitions)
	}
	if g.generation != nil {
		generation := *g.generation
		generation.Members = make(map[string]*Member, len(g.generation.Members))
		for id, m := range g.generation.Members {
			member := *m
			member.Partitions = make(map[string][]int, len(m.Partitions))
			for topic, partitions := range m.Partitions {
				member.Partitions[topic] = slices.Clone(partitions)
			}
			generation.Members[id] = &member
		}
		snapshot.Generation = &generation
	}
	return snapshot
}

func (g *Group) Commit(topic string, partition int, offset int64) {
	g.m.Lock()
	if g.commits == nil {
		g.commits = make(map[string]map[int]int64)
	}
	topicCommits, ok := g.commits[topic]
	if !ok {
		topicCommits = make(map[int]int64)
		g.commits[topic] = topicCommits
	}
	topicCommits[partition] = offset
	g.m.Unlock()

	log.Infof("kafka: group %v committed offset %v topic=%v partition=%v", g.Name, offset, topic, partition)
}

func (g *Group) Offset(topic string, partition int) int64 {
	if offset, ok := g.committedOffset(topic, partition); ok {
		return offset
	}
	// If there is no offset associated with a topic-partition under that consumer group the broker
	// does not set an error code (since it is not really an error), but returns empty metadata and sets the
	// offset field to -1.
	return -1
}

func (g *Group) committedOffset(topic string, partition int) (int64, bool) {
	g.m.RLock()
	defer g.m.RUnlock()

	offset, ok := g.commits[topic][partition]
	return offset, ok
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

func newMember(ctx *kafka.ClientContext, sessionTimeout int) *Member {
	return &Member{
		Partitions:     map[string][]int{},
		Client:         ctx,
		SessionTimeout: sessionTimeout,
	}
}

type groupMonitor struct {
	cluster string
	monitor *monitor.Kafka
}

func (gm *groupMonitor) LastRebalancing(group string, time time.Time) {
	gm.monitor.LastRebalancing.WithLabel(gm.cluster, group).Set(float64(time.Unix()))
}
