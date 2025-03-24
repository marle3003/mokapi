package store

import (
	"bufio"
	"bytes"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/kafka/joinGroup"
	"mokapi/kafka/syncGroup"
	"mokapi/providers/asyncapi3"
	"time"
)

type groupBalancer struct {
	group *Group
	join  chan joindata
	sync  chan syncdata
	stop  chan bool

	joins  []joindata
	config asyncapi3.BrokerBindings
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

func newGroupBalancer(group *Group, config asyncapi3.BrokerBindings) *groupBalancer {
	return &groupBalancer{
		group:  group,
		join:   make(chan joindata),
		sync:   make(chan syncdata),
		stop:   make(chan bool, 1),
		config: config,
	}
}

func (b *groupBalancer) Stop() {
	b.stop <- true
}

func (b *groupBalancer) run() {
	stop := make(chan bool, 1)
	var syncs []syncdata
	var assigns map[string]*groupAssignment
	prepareRebalance := func() {
		log.Infof("kafka: group %v state changed from %v to %v", b.group.Name, states[b.group.State], states[PreparingRebalance])
		// start a new rebalance
		b.group.State = PreparingRebalance
		b.joins = make([]joindata, 0)
		syncs = nil
		assigns = nil
		go b.finishJoin(stop)
	}

	timeoutMs := b.config.GroupMinSessionTimeoutMs
	if timeoutMs == 0 {
		timeoutMs = 6000
	}
	for {
		select {
		case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
			if b.group.Generation == nil {
				b.group.State = Empty
				continue
			}
			now := time.Now()
			for _, m := range b.group.Generation.Members {
				if m.Client.Heartbeat.Add(time.Duration(m.SessionTimeout) * time.Millisecond).Before(now) {
					log.Infof("kafka: consumer '%v' timed out in group '%v'", m.Client.ClientId, b.group.Name)
					prepareRebalance()
				}
			}
		case <-b.stop:
			stop <- true
			return
		case j := <-b.join:
			if b.group.State == CompletingRebalance {
				b.sendRebalanceInProgress(j.writer)
				continue
			}
			if b.group.State == Stable || b.group.State == Empty {
				prepareRebalance()
			}
			log.Infof("kafka: consumer '%v' is joining the group '%v'", j.client.ClientId, b.group.Name)
			b.joins = append(b.joins, j)
		case s := <-b.sync:
			switch {
			case s.assigns != nil: // leader sync
				assigns = s.assigns
				syncs = append(syncs, s)
				log.Infof("kafka: group %v state changed from %v to %v", b.group.Name, states[b.group.State], states[Stable])
				b.group.State = Stable
				for _, s := range syncs {
					memberName := s.client.Member[b.group.Name]
					assign := assigns[memberName]
					res := &syncGroup.Response{
						Assignment: assign.raw,
					}
					go b.respond(s.writer, res)
				}

				for memberName, assign := range assigns {
					for topicName, partitions := range assign.topics {
						b.group.Generation.Members[memberName].Partitions[topicName] = partitions
					}
				}

				log.Infof("kafka: received assignments from leader '%v' for group '%v'", s.client.ClientId, b.group.Name)
			case assigns == nil: // waiting for leader
				syncs = append(syncs, s)
			default:
				// we have leader sync and respond directly
				// a dead consumer from last generation will be kicked by heartbeat
				res := &syncGroup.Response{
					Assignment: assigns[s.client.Member[b.group.Name]].raw,
				}
				b.respond(s.writer, res)
			}
		}
	}
}

func (b *groupBalancer) finishJoin(stop chan bool) {
	rebalanceTimeoutMs := b.config.GroupInitialRebalanceDelayMs
	if rebalanceTimeoutMs == 0 {
		rebalanceTimeoutMs = 3000
	}
	// change to a better solution. If only one consumer is used, then we will wait the here too long when
	// consumer joins a second time
	//if b.group.Generation != nil {
	//	rebalanceTimeoutMs = b.group.Generation.RebalanceTimeoutMs
	//}

StopWaitingForConsumers:
	for {
		select {
		case <-stop:
			return
		case <-time.After(time.Duration(rebalanceTimeoutMs) * time.Millisecond):
			break StopWaitingForConsumers
		}
	}

	generation := b.group.NewGeneration()
	if len(b.joins) == 0 {
		log.Infof("kafka: group %v state changed from %v to %v", b.group.Name, states[b.group.State], states[Empty])
		b.group.State = Empty
		return
	}

	log.Infof("kafka group %v state changed from %v to %v", b.group.Name, states[b.group.State], states[CompletingRebalance])
	b.group.State = CompletingRebalance

	counter := map[string]*protocoldata{
		"": {counter: -1},
	}
	var protocol string
	for _, j := range b.joins {
		memberId := j.client.GetOrCreateMemberId(b.group.Name)
		generation.Members[memberId] = newMember(j.client, j.sessionTimeout)

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
	generation.LeaderId = leader.client.Member[b.group.Name]
	generation.RebalanceTimeoutMs = leader.rebalanceTimeout
	members := make([]joinGroup.Member, 0, len(b.joins))
	members = append(members, joinGroup.Member{
		MemberId: generation.LeaderId,
		MetaData: counter[protocol].metadata[generation.LeaderId],
	})

	for _, j := range b.joins[1:] {
		memberId := j.client.Member[b.group.Name]
		members = append(members, joinGroup.Member{
			MemberId: memberId,
			MetaData: counter[protocol].metadata[memberId],
		})
		go b.respond(j.writer, &joinGroup.Response{
			GenerationId: int32(generation.Id),
			Leader:       generation.LeaderId,
			MemberId:     memberId,
			ProtocolName: protocol,
		})
	}

	go b.respond(leader.writer, &joinGroup.Response{
		GenerationId: int32(generation.Id),
		Leader:       generation.LeaderId,
		MemberId:     generation.LeaderId,
		ProtocolName: protocol,
		Members:      members,
	})
}

func (b *groupBalancer) sendRebalanceInProgress(w kafka.ResponseWriter) {
	go b.respond(w, &joinGroup.Response{ErrorCode: kafka.RebalanceInProgress})
}

func (b *groupBalancer) respond(w kafka.ResponseWriter, msg kafka.Message) {
	go func() {
		err := w.Write(msg)
		if err != nil {
			log.Errorf("kafka group balancer for group %v: %v", b.group.Name, err)
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
