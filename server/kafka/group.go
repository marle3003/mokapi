package kafka

import (
	log "github.com/sirupsen/logrus"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/joinGroup"
	"mokapi/server/kafka/protocol/syncGroup"
	"time"
)

type groupState int

var (
	empty               groupState = 0
	preparingRebalance  groupState = 1
	completingRebalance groupState = 2
	// The rebalancing already took place and consumers are happily consuming
	stable groupState = 3
)

type group struct {
	name string
	// leader is at index 0, followed by the followers
	members     []groupMember
	coordinator broker
	// Upon every completion of the join group phase, the coordinator
	// increments a GenerationId for the group. This is returned as a field
	// in the response to each member, and is sent in heartbeats and offset
	// commit requests. When the coordinator rebalances a group, the coordinator
	// will send an error code indicating that the member needs to rejoin. If the
	// member does not rejoin before a rebalance completes, then it will have an
	// old generationId, which will cause ILLEGAL_GENERATION errors when included
	// in new requests.
	generationId     int
	state            groupState
	balancer         *groupBalancer
	rebalanceTimeout int
	sessionTimeout   int
}

type groupMember struct {
	consumer *client
	memberId string
}

type groupBalancer struct {
	g              *group
	join           chan join
	sync           chan syncData
	stop           chan bool
	rebalanceDelay int
}

type strategyCounter struct {
	name    string
	counter int
}

type join struct {
	consumer *client
	write    func(protocol.Message)
	// assignment strategy, ordered by consumer's preference
	protocols        []groupAssignmentStrategy
	generationId     int
	rebalanceTimeout int
	sessionTimeout   int
}

func (b *groupBalancer) startGroupWatcher() {
	ticker := time.NewTicker(time.Duration(b.g.sessionTimeout) * time.Millisecond)
	for {
		select {
		case <-b.stop:
			return
		case <-ticker.C:
			if b.g.state != stable {
				continue
			}
			i := 0
			t := time.Now()
			timeout := int64(b.g.sessionTimeout)
			needRebalance := false
			for _, m := range b.g.members {

				d := t.Sub(m.consumer.lastHeartbeat)
				if d.Milliseconds() < timeout {
					b.g.members[i] = m
					i++
				} else {
					needRebalance = true
					log.Infof("kafka: session timeout of consumer %q in group %q", m.consumer.id, b.g.name)
				}
			}

			b.g.members = b.g.members[:i]

			if needRebalance {
				if len(b.g.members) == 0 {
					b.g.state = empty
				} else {
					b.g.state = preparingRebalance
					go b.startJoin()
				}
			}
		}
	}
}

func (j join) getMetadata(strategy string) []byte {
	for _, p := range j.protocols {
		if p.assignmentStrategy == strategy {
			return p.metadata
		}
	}
	return nil
}

type syncData struct {
	consumer     *client
	assignments  map[string][]byte
	write        func(protocol.Message)
	generationId int
}

func newGroupBalancer(g *group, rebalanceDelay int) *groupBalancer {
	return &groupBalancer{
		g:              g,
		join:           make(chan join),
		sync:           make(chan syncData),
		stop:           make(chan bool),
		rebalanceDelay: rebalanceDelay,
	}
}

func (b *groupBalancer) startSync() {
	members := make([]syncData, 0)
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
				if len(members) == len(b.g.members) {
					break StopWaitingForConsumers
				}
			} else {
				r := &syncGroup.Response{
					ErrorCode: 22, // IllegalGenerationCode
				}
				s.write(r)
			}
		case <-time.After(time.Duration(b.g.rebalanceTimeout) * time.Millisecond):
			b.g.state = preparingRebalance
			for _, m := range members {
				r := &syncGroup.Response{
					ErrorCode: 27, // REBALANCE_IN_PROGRESS
				}
				m.write(r)
			}
			return
		}
	}

	for _, m := range members {
		r := &syncGroup.Response{
			ErrorCode:  0,
			Assignment: assigments[m.consumer.id],
		}
		m.write(r)
		m.consumer.group = b.g
	}

	b.g.state = stable
	b.startGroupWatcher()
}

func (b *groupBalancer) startJoin() {

	members := make([]join, 0)
StopWaitingForConsumers:
	for {
		select {
		case <-b.stop:
			return
		case j := <-b.join:
			members = append(members, j)
		case <-time.After(5 * time.Second):
			break StopWaitingForConsumers
		}
	}

	if len(members) == 0 {
		b.g.state = empty
		return
	}

	// switch to syncing state
	b.g.state = completingRebalance

	i := 0
	for _, m := range b.g.members {
		if hasJoined(m.consumer, members) {
			b.g.members[i] = m
			i++
		}
	}
	b.g.members = b.g.members[:i]

	strategies := make([]strategyCounter, 0)
	for _, m := range members {
		if !isMember(m.consumer, b.g.members) {
			b.g.members = append(b.g.members, groupMember{consumer: m.consumer})
		}
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
		if s.counter == len(b.g.members) {
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
		Leader:       b.g.members[0].consumer.id,
		MemberId:     b.g.members[0].consumer.id,
		Members:      make([]joinGroup.Member, 0, len(b.g.members)),
	}

	send := make([]func(), 0, len(b.g.members))
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
			b.g.sessionTimeout = m.sessionTimeout
			b.g.rebalanceTimeout = m.rebalanceTimeout
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

func isMember(c *client, members []groupMember) bool {
	for _, m := range members {
		if m.consumer.id == c.id {
			return true
		}
	}
	return false
}

func hasJoined(c *client, members []join) bool {
	for _, m := range members {
		if m.consumer.id == c.id {
			return true
		}
	}
	return false
}
