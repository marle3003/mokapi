package kafka

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"math"
	"mokapi/providers/utils"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/apiVersion"
	"mokapi/server/kafka/protocol/fetch"
	"mokapi/server/kafka/protocol/findCoordinator"
	"mokapi/server/kafka/protocol/heartbeat"
	"mokapi/server/kafka/protocol/joinGroup"
	"mokapi/server/kafka/protocol/listOffsets"
	"mokapi/server/kafka/protocol/metaData"
	"mokapi/server/kafka/protocol/offsetCommit"
	"mokapi/server/kafka/protocol/offsetFetch"
	"mokapi/server/kafka/protocol/produce"
	"mokapi/server/kafka/protocol/syncGroup"
	"net"
	"time"
)

func (s *Binding) handle(conn net.Conn) {
	defer func() {
		log.Info("Closing kafka connection")
		conn.Close()
	}()

	for {
		h, msg, err := protocol.ReadMessage(conn)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Error(err)
			return
		}

		c, exists := s.clients[h.ClientId]
		if !exists {
			c = &client{id: h.ClientId}
			s.clients[h.ClientId] = c
		}
		c.lastHeartbeat = time.Now()

		switch h.ApiKey {
		case protocol.ApiVersions:
			r := s.processApiVersion()
			protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, r)
		case protocol.Metadata:
			r := s.processMetadata(msg.(*metaData.Request))
			protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, r)
		case protocol.FindCoordinator:
			r := s.processFindCoordinator(msg.(*findCoordinator.Request))
			protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, r)
		case protocol.JoinGroup:
			errorCode := s.processJoinGroup(h, msg.(*joinGroup.Request), conn)
			if errorCode != 0 {
				protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, &joinGroup.Response{ErrorCode: errorCode})
			}
		case protocol.SyncGroup:
			s.handleSyncGroup(h, msg.(*syncGroup.Request), conn)
		case protocol.OffsetFetch:
			r := s.processOffSetFetch(msg.(*offsetFetch.Request))
			protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, r)
		case protocol.Fetch:
			if c.group != nil && c.group.state != stable {
				protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, &fetch.Response{ErrorCode: protocol.RebalanceInProgress})
			} else {
				r := s.processFetch(msg.(*fetch.Request))
				protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, r)
			}
		case protocol.Heartbeat:
			protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, &heartbeat.Response{})
		case protocol.Produce:
			r := msg.(*produce.Request)
			for _, t := range r.Topics {
				topic := s.topics[t.Name]
				topic.addRecords(int(t.Data.Partition), t.Data.Record.Batches)
			}
		case protocol.ListOffsets:
			r := s.processListOffsets(msg.(*listOffsets.Request))
			protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, r)
		case protocol.OffsetCommit:
			r := s.processOffsetCommit(msg.(*offsetCommit.Request))
			protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, r)
		}
	}
}

func (s *Binding) processListOffsets(req *listOffsets.Request) *listOffsets.Response {
	r := &listOffsets.Response{Topics: make([]listOffsets.ResponseTopic, 0)}

	for _, rt := range req.Topics {
		if t, ok := s.topics[rt.Name]; ok {
			partitions := make([]listOffsets.ResponsePartition, 0)
			for _, rp := range rt.Partitions {
				p := t.partitions[int(rp.Index)]
				part := listOffsets.ResponsePartition{
					Index:     rp.Index,
					ErrorCode: 0,
					Timestamp: -1,
					Offset:    p.offset,
				}

				if rp.Timestamp == -2 { // first offset
					part.Offset = p.startOffset
				} else if rp.Timestamp == -1 { // lastOffset
					part.Offset = p.offset
				}

				//if part.Offset >= 0 {
				//	part.Timestamp = protocol.Timestamp(p.segments[p.activeSegment].log[0].Records[0].Time)
				//}

				partitions = append(partitions, part)
			}
			r.Topics = append(r.Topics, listOffsets.ResponseTopic{
				Name:       rt.Name,
				Partitions: partitions,
			})
		}
	}

	return r
}

func (s *Binding) processApiVersion() *apiVersion.Response {
	r := &apiVersion.Response{
		ApiKeys: make([]apiVersion.ApiKeyResponse, 0, len(protocol.ApiTypes)),
	}
	for k, t := range protocol.ApiTypes {
		r.ApiKeys = append(r.ApiKeys, apiVersion.NewApiKeyResponse(k, t))
	}
	return r
}

func (s *Binding) processMetadata(req *metaData.Request) *metaData.Response {
	r := &metaData.Response{
		Brokers:   make([]metaData.ResponseBroker, 0, len(s.brokers)),
		Topics:    make([]metaData.ResponseTopic, 0, len(req.Topics)),
		ClusterId: "mokapi",
	}

	for _, b := range s.brokers {
		r.Brokers = append(r.Brokers, metaData.ResponseBroker{
			NodeId: int32(b.id),
			Host:   b.host,
			Port:   int32(b.port),
		})
	}

	r.ControllerId = r.Brokers[0].NodeId // using first broker as controller

	if len(req.Topics) == 0 {
		for n, t := range s.topics {
			resT := metaData.ResponseTopic{
				Name:       n,
				Partitions: make([]metaData.ResponsePartition, 0, len(t.partitions)),
			}

			for i, p := range t.partitions {
				resT.Partitions = append(resT.Partitions, metaData.ResponsePartition{
					PartitionIndex: int32(i),
					LeaderId:       int32(p.leader.id),
					ReplicaNodes:   []int32{1},
					IsrNodes:       []int32{1},
				})
			}

			r.Topics = append(r.Topics, resT)
		}
		return r
	}

	for _, reqT := range req.Topics {
		if t, ok := s.topics[reqT.Name]; ok {
			resT := metaData.ResponseTopic{
				Name:       reqT.Name,
				Partitions: make([]metaData.ResponsePartition, 0, len(t.partitions)),
			}

			for i, p := range t.partitions {
				resT.Partitions = append(resT.Partitions, metaData.ResponsePartition{
					PartitionIndex: int32(i),
					LeaderId:       int32(p.leader.id),
					ReplicaNodes:   []int32{1},
					IsrNodes:       []int32{1},
				})
			}

			r.Topics = append(r.Topics, resT)
		} else {
			r.Topics = append(r.Topics, metaData.ResponseTopic{
				ErrorCode: protocol.UnknownTopicOrPartition,
				Name:      reqT.Name,
			})
		}
	}

	return r

}

func (s *Binding) processFindCoordinator(req *findCoordinator.Request) *findCoordinator.Response {
	r := &findCoordinator.Response{}

	switch req.KeyType {
	case 0: // group
		var g *group
		if e, ok := s.groups[req.Key]; ok {
			g = e
		} else {
			g = &group{
				name:        req.Key,
				coordinator: s.brokers[0],
			}
			g.balancer = newGroupBalancer(g, s.kafka.Group.Initial.Rebalance.Delay)
			s.groups[req.Key] = g
		}

		r.NodeId = int32(g.coordinator.id)
		r.Host = g.coordinator.host
		r.Port = int32(g.coordinator.port)
	default:
		msg := fmt.Sprintf("unsupported key type '%v' in find coordinator request", req.KeyType)
		log.Error(msg)
		r.ErrorCode = -1
		r.ErrorMessage = msg
		return r
	}
	return r
}

func (s *Binding) processJoinGroup(h *protocol.Header, req *joinGroup.Request, w io.Writer) protocol.ErrorCode {
	var g *group
	var exists bool
	if g, exists = s.groups[req.GroupId]; !exists {
		return protocol.InvalidGroupId
	} else if g.state == completingRebalance {
		return protocol.RebalanceInProgress
	} else if g.state == empty || g.state == stable {
		g.state = preparingRebalance
		go g.balancer.startJoin()
	}

	if len(req.MemberId) == 0 {
		memberId := fmt.Sprintf("%v-%v", h.ClientId, utils.NewGuid())
		protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, &joinGroup.Response{
			ErrorCode: 79, // MEMBER_ID_REQUIRED
			MemberId:  memberId,
		})
		return 0
	}

	j := join{
		consumer:  s.clients[h.ClientId],
		protocols: make([]groupAssignmentStrategy, 0, len(req.Protocols)),
		write: func(msg protocol.Message) {
			protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, msg)
		},
		rebalanceTimeout: int(req.RebalanceTimeoutMs),
		sessionTimeout:   int(req.SessionTimeoutMs),
	}

	for _, p := range req.Protocols {
		j.protocols = append(j.protocols, groupAssignmentStrategy{
			assignmentStrategy: p.Name,
			metadata:           p.MetaData,
		})
	}

	g.balancer.join <- j

	return protocol.None
}

func (s *Binding) handleSyncGroup(h *protocol.Header, req *syncGroup.Request, w io.Writer) int {
	var g *group
	var exists bool
	if g, exists = s.groups[req.GroupId]; !exists {
		return -1
	}

	if g.state == preparingRebalance {
		return -27 // REBALANCE_IN_PROGRESS
	}

	sync := syncData{
		consumer:     s.clients[h.ClientId],
		generationId: int(req.GenerationId),
		write: func(msg protocol.Message) {
			protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, msg)
		},
	}

	if req.GroupAssignments != nil {
		sync.assignments = make(map[string][]byte)
		for _, a := range req.GroupAssignments {
			sync.assignments[a.MemberId] = a.Assignment
		}
	}

	g.balancer.sync <- sync

	return 0
}

func (s *Binding) processOffSetFetch(req *offsetFetch.Request) *offsetFetch.Response {
	r := &offsetFetch.Response{
		Topics: make([]offsetFetch.ResponseTopic, 0, len(req.Topics)),
	}

	// currently offset is not separated by groups
	for _, rt := range req.Topics {
		t := s.topics[rt.Name]
		resTopic := offsetFetch.ResponseTopic{Name: rt.Name, Partitions: make([]offsetFetch.Partition, 0, len(rt.PartitionIndexes))}
		for _, rp := range rt.PartitionIndexes {
			p := t.partitions[int(rp)]
			resTopic.Partitions = append(resTopic.Partitions, offsetFetch.Partition{
				Index:           rp,
				CommittedOffset: p.getOffset(req.GroupId),
			})
		}
		r.Topics = append(r.Topics, resTopic)
	}

	return r
}

func (s *Binding) processFetch(req *fetch.Request) *fetch.Response {
	r := &fetch.Response{Topics: make([]fetch.ResponseTopic, 0)}

	start := time.Now().Add(time.Duration(req.MaxWaitMs-200) * time.Millisecond) // -200: working load time
	size := int32(0)
	for {
		// currently offset is not separated by groups
		for _, rt := range req.Topics {
			t := s.topics[rt.Name]
			resTopic := fetch.ResponseTopic{Name: rt.Name, Partitions: make([]fetch.ResponsePartition, 0, len(rt.Partitions))}
			for _, rp := range rt.Partitions {
				p := t.partitions[int(rp.Index)]
				resPar := fetch.ResponsePartition{
					Index:                rp.Index,
					HighWatermark:        p.offset,
					LastStableOffset:     p.offset,
					LogStartOffset:       0,
					PreferredReadReplica: -1,
				}

				record, recordSize := p.read(rp.FetchOffset, rp.MaxBytes)
				resPar.RecordSet = record
				size += recordSize
				resTopic.Partitions = append(resTopic.Partitions, resPar)
			}
			r.Topics = append(r.Topics, resTopic)
		}

		if time.Now().After(start) || size > req.MinBytes {
			return r
		}

		time.Sleep(time.Duration(math.Floor(0.2*float64(req.MaxWaitMs))) * time.Millisecond)
	}
}

func (s *Binding) processOffsetCommit(req *offsetCommit.Request) *offsetCommit.Response {
	r := &offsetCommit.Response{
		Topics: make([]offsetCommit.ResponseTopic, 0, len(req.Topics)),
	}

	// currently offset is not separated by groups
	for _, rt := range req.Topics {
		t := s.topics[rt.Name]
		resTopic := offsetCommit.ResponseTopic{Name: rt.Name, Partitions: make([]offsetCommit.ResponsePartition, 0, len(rt.Partitions))}
		for _, rp := range rt.Partitions {
			p := t.partitions[int(rp.Index)]
			p.setOffset(req.GroupId, rp.Offset)
			resTopic.Partitions = append(resTopic.Partitions, offsetCommit.ResponsePartition{
				Index: rp.Index,
			})
		}
		r.Topics = append(r.Topics, resTopic)
	}

	return r
}
