package kafka

import (
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/joinGroup"
	"mokapi/server/kafka/protocol/syncGroup"
	"time"
)

type groupState int

var (
	empty   groupState = 0
	joining groupState = 1
	syncing groupState = 2
	// The rebalancing already took place and consumers are happily consuming
	stable groupState = 3
)

type group struct {
	// leader is at index 0, followed by the followers
	consumers   []consumer
	coordinator broker
	// Upon every completion of the join group phase, the coordinator
	// increments a GenerationId for the group. This is returned as a field
	// in the response to each member, and is sent in heartbeats and offset
	// commit requests. When the coordinator rebalances a group, the coordinator
	// will send an error code indicating that the member needs to rejoin. If the
	// member does not rejoin before a rebalance completes, then it will have an
	// old generationId, which will cause ILLEGAL_GENERATION errors when included
	// in new requests.
	generationId int
	state        groupState
	balancer     *groupBalancer
}

type groupBalancer struct {
	g    *group
	join chan join
	sync chan sync
}

type strategyCounter struct {
	name    string
	counter int
}

type join struct {
	consumer consumer
	write    func(protocol.Message)
	// assignment strategy, ordered by consumer's preference
	protocols    []groupAssignmentStrategy
	generationId int
}

func (j join) getMetadata(strategy string) []byte {
	for _, p := range j.protocols {
		if p.assignmentStrategy == strategy {
			return p.metadata
		}
	}
	return nil
}

type sync struct {
	consumer     consumer
	assignments  map[string][]byte
	write        func(protocol.Message)
	generationId int
}

func newGroupBalancer(g *group) *groupBalancer {
	return &groupBalancer{
		g:    g,
		join: make(chan join),
		sync: make(chan sync),
	}
}

func (b *groupBalancer) startSync() {
	members := make([]sync, 0)
	assigments := make(map[string][]byte)

StopWaitingForConsumers:
	for {
		select {
		case s := <-b.sync:
			members = append(members, s)
			if b.g.generationId == s.generationId {

				if s.assignments != nil {
					assigments = s.assignments
				}
				if len(members) == len(b.g.consumers) {
					break StopWaitingForConsumers
				}
			} else {
				r := &syncGroup.Response{
					ErrorCode: 22, // IllegalGenerationCode
				}
				s.write(r)
			}
		case <-time.After(2 * time.Second):
			panic("timeout") // todo
		}
	}

	for _, m := range members {
		r := &syncGroup.Response{
			ErrorCode:  0,
			Assignment: assigments[m.consumer.id],
		}
		m.write(r)
	}

	b.g.state = stable
}

func (b *groupBalancer) startJoin() {

	members := make([]join, 0)
StopWaitingForConsumers:
	for {
		select {
		case j := <-b.join:
			members = append(members, j)
		case <-time.After(2 * time.Second):
			break StopWaitingForConsumers
		}
	}

	// switch to syncing state
	b.g.state = syncing

	strategies := make([]strategyCounter, 0)
	for _, m := range members {
		b.g.consumers = append(b.g.consumers, m.consumer)
		for _, p := range m.protocols {
			shouldAdd := true
			for _, s := range strategies {
				if s.name == p.assignmentStrategy {
					s.counter++
					shouldAdd = false
				}
			}
			if shouldAdd {
				strategies = append(strategies, strategyCounter{name: p.assignmentStrategy, counter: 1})
			}
		}
	}

	chosenStrategy := ""
	for _, s := range strategies {
		if s.counter == len(b.g.consumers) {
			chosenStrategy = s.name
			break
		}
	}

	if len(chosenStrategy) == 0 {
		// todo error handling
	}

	rLeader := &joinGroup.Response{
		GenerationId: int32(b.g.generationId),
		ProtocolName: chosenStrategy,
		Leader:       b.g.consumers[0].id,
		MemberId:     b.g.consumers[0].id,
		Members:      make([]joinGroup.Member, 0, len(b.g.consumers)),
	}

	send := make([]func(), 0, len(b.g.consumers))
	for i, m := range members {
		if i > 0 {
			r := &joinGroup.Response{
				GenerationId: -1,
				ProtocolName: chosenStrategy,
				Leader:       members[0].consumer.id,
				MemberId:     m.consumer.id,
			}
			send = append(send, func() { m.write(r) })
		} else {
			send = append(send, func() { m.write(rLeader) })
		}
		rLeader.Members = append(rLeader.Members, joinGroup.Member{
			MemberId:        m.consumer.id,
			GroupInstanceId: "",
			MetaData:        m.getMetadata(chosenStrategy),
		})
	}

	for _, s := range send {
		s()
	}

	b.startSync()
}
