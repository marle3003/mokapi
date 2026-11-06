package store

import (
	"mokapi/kafka"
	"mokapi/kafka/joinGroup"
	"mokapi/kafka/syncGroup"
	"mokapi/providers/asyncapi3"
	"slices"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type groupBalancer struct {
	group *Group
	join  chan joindata
	sync  chan syncdata
	stop  chan bool

	joinsMu sync.Mutex
	joins   []joindata
	config  asyncapi3.BrokerBindings
	monitor *groupMonitor
}

type joindata struct {
	client           *kafka.ClientContext
	writer           kafka.ResponseWriter
	protocolType     string
	protocols        []joinGroup.Protocol
	rebalanceTimeout int
	sessionTimeout   int
	log              func(res any)
}

type syncdata struct {
	client       *kafka.ClientContext
	writer       kafka.ResponseWriter
	generationId int32
	protocolType string
	protocolName string
	assigns      map[string]*groupAssignment
	log          func(res any)
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

func newGroupBalancer(group *Group, config asyncapi3.BrokerBindings, monitor *groupMonitor) *groupBalancer {
	return &groupBalancer{
		group:   group,
		join:    make(chan joindata),
		sync:    make(chan syncdata),
		stop:    make(chan bool, 1),
		config:  config,
		monitor: monitor,
	}
}

// joins are appended by the run loop and read by finishJoin on another goroutine.
func (b *groupBalancer) resetJoins() {
	b.joinsMu.Lock()
	defer b.joinsMu.Unlock()
	b.joins = make([]joindata, 0)
}

func (b *groupBalancer) addJoin(j joindata) {
	b.joinsMu.Lock()
	defer b.joinsMu.Unlock()
	b.joins = append(b.joins, j)
}

func (b *groupBalancer) joinsSnapshot() []joindata {
	b.joinsMu.Lock()
	defer b.joinsMu.Unlock()
	return slices.Clone(b.joins)
}

func (b *groupBalancer) Stop() {
	b.stop <- true
}

func (b *groupBalancer) run() {
	stop := make(chan bool, 1)
	var syncs []syncdata
	var assigns map[string]*groupAssignment
	prepareRebalance := func() {
		log.Infof("kafka: group %v state changed from %v to %v", b.group.Name, states[b.group.State()], states[PreparingRebalance])
		// start a new rebalance
		b.group.SetState(PreparingRebalance)
		b.resetJoins()
		syncs = nil
		assigns = nil
		go b.finishJoin(stop)
	}

	timeoutMs := b.config.GroupMinSessionTimeoutMs
	if timeoutMs == 0 {
		timeoutMs = 6000
	}
	gracePeriod := time.Duration(2) * time.Second
	for {
		select {
		case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
			generation := b.group.Generation()
			if generation == nil {
				b.group.SetState(Empty)
				continue
			}
			if b.group.State() != Stable {
				continue
			}
			now := time.Now().Add(gracePeriod)
			for _, m := range generation.Members {
				if m.Client.Heartbeat().Add(time.Duration(m.SessionTimeout) * time.Millisecond).Before(now) {
					log.Infof("kafka: consumer '%v' timed out in group '%v': last heartbeat %v", m.Client.ClientId(), b.group.Name, m.Client.Heartbeat().Format(time.RFC822))
					prepareRebalance()
				}
			}
		case <-b.stop:
			stop <- true
			return
		case j := <-b.join:
			state := b.group.State()
			if state == CompletingRebalance {
				b.sendRebalanceInProgress(j.writer)
				continue
			}
			if state == Stable || state == Empty {
				prepareRebalance()
			}
			log.Infof("kafka: consumer '%v' is joining the group '%v' with session timeout %v", j.client.ClientId(), b.group.Name, j.sessionTimeout)
			b.addJoin(j)
		case s := <-b.sync:
			switch {
			case s.assigns != nil: // leader sync
				assigns = s.assigns
				syncs = append(syncs, s)
				log.Infof("kafka: group %v state changed from %v to %v", b.group.Name, states[b.group.State()], states[Stable])
				b.group.SetState(Stable)
				for memberName, assign := range assigns {
					b.group.AssignPartitions(memberName, assign.topics)
				}

				if b.monitor != nil {
					b.monitor.LastRebalancing(b.group.Name, time.Now())
				}

				log.Infof("kafka: received assignments from leader '%v' for group '%v'", s.client.ClientId(), b.group.Name)

				for _, sync := range syncs {
					memberName := sync.client.Member[b.group.Name]
					assign := assigns[memberName]
					res := &syncGroup.Response{
						ProtocolType: sync.protocolType,
						ProtocolName: sync.protocolName,
						Assignment:   assign.raw,
					}
					go b.respond(sync.writer, res)
					go func() {
						sync.log(newKafkaSyncGroupResponse(res, assign))
					}()
				}

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

	joins := b.joinsSnapshot()
	generation := b.group.NewGeneration()
	if len(joins) == 0 {
		b.group.SetGeneration(generation)
		log.Infof("kafka: group %v state changed from %v to %v", b.group.Name, states[b.group.State()], states[Empty])
		b.group.SetState(Empty)
		return
	}

	counter := map[string]*protocoldata{
		"": {counter: -1},
	}
	var protocol string
	for _, j := range joins {
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

	leader := joins[0]
	generation.LeaderId = leader.client.Member[b.group.Name]
	generation.RebalanceTimeoutMs = leader.rebalanceTimeout
	// publish the complete generation before any member learns about it
	b.group.SetGeneration(generation)
	log.Infof("kafka group %v state changed from %v to %v", b.group.Name, states[b.group.State()], states[CompletingRebalance])
	b.group.SetState(CompletingRebalance)

	members := make([]joinGroup.Member, 0, len(joins))
	members = append(members, joinGroup.Member{
		MemberId: generation.LeaderId,
		MetaData: counter[protocol].metadata[generation.LeaderId],
	})

	for _, j := range joins[1:] {
		memberId := j.client.Member[b.group.Name]
		members = append(members, joinGroup.Member{
			MemberId: memberId,
			MetaData: counter[protocol].metadata[memberId],
		})
		res := &joinGroup.Response{
			GenerationId: int32(generation.Id),
			Leader:       generation.LeaderId,
			MemberId:     memberId,
			ProtocolType: j.protocolType,
			ProtocolName: protocol,
		}
		go b.respond(j.writer, res)
		go func() {
			j.log(newKafkaJoinGroupResponse(res))
		}()
	}

	res := &joinGroup.Response{
		GenerationId: int32(generation.Id),
		Leader:       generation.LeaderId,
		MemberId:     generation.LeaderId,
		ProtocolType: leader.protocolType,
		ProtocolName: protocol,
		Members:      members,
	}
	go b.respond(leader.writer, res)
	go func() {
		leader.log(newKafkaJoinGroupResponse(res))
	}()
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

func newKafkaJoinGroupResponse(res *joinGroup.Response) *KafkaJoinGroupResponse {
	r := &KafkaJoinGroupResponse{
		GenerationId: res.GenerationId,
		ProtocolName: res.ProtocolName,
		MemberId:     res.MemberId,
		LeaderId:     res.Leader,
	}
	for _, m := range res.Members {
		r.Members = append(r.Members, m.MemberId)
	}

	return r
}

func newKafkaSyncGroupResponse(res *syncGroup.Response, assign *groupAssignment) *KafkaSyncGroupResponse {
	return &KafkaSyncGroupResponse{
		ProtocolType: res.ProtocolType,
		ProtocolName: res.ProtocolName,
		Assignment: KafkaSyncGroupAssignment{
			Version: assign.version,
			Topics:  assign.topics,
		},
	}
}
