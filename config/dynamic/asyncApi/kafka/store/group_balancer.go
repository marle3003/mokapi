package store

import (
	"bufio"
	"bytes"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/kafka/joinGroup"
	"mokapi/kafka/syncGroup"
	"time"
)

type groupBalancerNew struct {
	group *Group
	join  chan joindata
	sync  chan syncdata
	stop  chan bool

	joins []joindata
}

type joindata struct {
	client           *kafka.ClientContext
	writer           kafka.ResponseWriter
	protocols        []joinGroup.Protocol
	rebalanceTimeout int
	sessionTimeout   int
}

type syncdata struct {
	client  *kafka.ClientContext
	writer  kafka.ResponseWriter
	assigns map[string]*groupAssignment
}

type protocoldata struct {
	counter  int
	metadata map[string][]byte
}

type groupAssignment struct {
	version  int16
	topics   map[string][]int
	userData []byte
	raw      []byte
}

func newGroupBalancerNew(group *Group) *groupBalancerNew {
	return &groupBalancerNew{
		group: group,
		join:  make(chan joindata),
		sync:  make(chan syncdata),
		stop:  make(chan bool, 1),
	}
}

func (b *groupBalancerNew) Stop() {
	b.stop <- true
}

func (b *groupBalancerNew) run() {
	stop := make(chan bool, 1)
	var syncs []syncdata
	var assigns map[string]*groupAssignment
	for {
		select {
		case <-b.stop:
			stop <- true
			return
		case j := <-b.join:
			if b.group.State == Stable {
				// start a new rebalance
				b.group.State = Joining
				b.joins = make([]joindata, 0)
				syncs = nil
				assigns = nil
				go b.finishJoin(stop)
			}
			b.joins = append(b.joins, j)
		case s := <-b.sync:
			switch {
			case s.assigns != nil: // leader sync
				assigns = s.assigns
				syncs = append(syncs, s)
				b.group.State = Stable
				for _, s := range syncs {
					res := &syncGroup.Response{
						Assignment: assigns[s.client.Member[b.group.Name]].raw,
					}
					go b.respond(s.writer, res)
				}
			case assigns == nil: // waiting for leader
				syncs = append(syncs, s)
			default:
				// we have leader sync and respond directly
				// a dead consumer will be kicked by heartbeat
				res := &syncGroup.Response{
					Assignment: assigns[s.client.Member[b.group.Name]].raw,
				}
				b.respond(s.writer, res)
			}
		}
	}
}

func (b *groupBalancerNew) finishJoin(stop chan bool) {
StopWaitingForConsumers:
	for {
		select {
		case <-stop:
			return
		case <-time.After(time.Duration(3000) * time.Millisecond):
			break StopWaitingForConsumers
		}
	}

	b.group.State = AwaitingSync
	generation := b.group.NewGeneration()

	counter := map[string]*protocoldata{
		"": {counter: -1},
	}
	var protocol string
	for _, j := range b.joins {
		memberId := j.client.GetOrCreateMemberId(b.group.Name)
		generation.Members[memberId] = &Member{}

		for _, proto := range j.protocols {
			if _, ok := counter[proto.Name]; !ok {
				counter[proto.Name] = &protocoldata{metadata: make(map[string][]byte)}
			}
			p := counter[proto.Name]
			p.counter++
			p.metadata[memberId] = proto.MetaData
			if counter[proto.Name].counter > counter[protocol].counter {
				protocol = proto.Name
			}
		}

	}

	generation.Protocol = protocol

	leader := b.joins[0]
	leaderMemberId := leader.client.Member[b.group.Name]
	members := make([]joinGroup.Member, 0, len(b.joins))
	members = append(members, joinGroup.Member{
		MemberId: leaderMemberId,
		MetaData: counter[protocol].metadata[leaderMemberId],
	})

	for _, j := range b.joins[1:] {
		memberId := j.client.Member[b.group.Name]
		members = append(members, joinGroup.Member{
			MemberId: memberId,
			MetaData: counter[protocol].metadata[memberId],
		})
		go b.respond(j.writer, &joinGroup.Response{
			GenerationId: int32(generation.Id),
			Leader:       leaderMemberId,
			MemberId:     memberId,
			ProtocolName: protocol,
		})
	}

	go b.respond(leader.writer, &joinGroup.Response{
		GenerationId: int32(generation.Id),
		Leader:       leaderMemberId,
		MemberId:     leaderMemberId,
		ProtocolName: protocol,
		Members:      members,
	})
}

func (b *groupBalancerNew) respond(w kafka.ResponseWriter, msg kafka.Message) {
	go func() {
		err := w.Write(msg)
		if err != nil {
			log.Errorf("kafka group balancer: %v", err)
		}
	}()
}

func newGroupAssignment(b []byte) *groupAssignment {
	g := &groupAssignment{}
	g.raw = b
	r := bufio.NewReader(bytes.NewReader(b))
	d := kafka.NewDecoder(r, len(b))
	g.version = d.ReadInt16()

	g.topics = make(map[string][]int)
	n := int(d.ReadInt32())
	for i := 0; i < n; i++ {
		key := d.ReadString()
		value := make([]int, 0)

		nPartition := int(d.ReadInt32())
		for j := 0; j < nPartition; j++ {
			index := d.ReadInt32()
			value = append(value, int(index))
		}
		g.topics[key] = value
	}

	g.userData = d.ReadBytes()

	return g
}
