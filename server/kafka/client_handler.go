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
		err := conn.Close()
		if err != nil {
			log.Errorf("unable to close kafka connection")
		}
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

		s.clientsMutex.Lock()
		c, exists := s.clients[h.ClientId]
		if !exists {
			c = &client{id: h.ClientId}
			s.clients[h.ClientId] = c
		}
		s.clientsMutex.Unlock()
		c.lastHeartbeat = time.Now()

		log.Infof("API %v, CorrelationID %v", h.ApiKey, h.CorrelationId)

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
			errorCode := s.processJoinGroup(h, msg.(*joinGroup.Request), c, conn)
			if errorCode != 0 {
				protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, &joinGroup.Response{ErrorCode: errorCode})
			}
		case protocol.SyncGroup:
			errorCode := s.handleSyncGroup(h, msg.(*syncGroup.Request), c, conn)
			if errorCode != 0 {
				protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, &syncGroup.Response{ErrorCode: errorCode})
			}
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
			if c.group != nil && c.group.state != stable {
				protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, &fetch.Response{ErrorCode: protocol.RebalanceInProgress})
			} else if c.group == nil {
				protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, &fetch.Response{ErrorCode: protocol.RebalanceInProgress})
			} else {
				protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, &heartbeat.Response{})
			}
		case protocol.Produce:
			// todo
			r := msg.(*produce.Request)
			res := &produce.Response{}
			for _, t := range r.Topics {
				topic := s.topics[t.Name]
				err := topic.addRecords(int(t.Data.Partition), t.Data.Record)
				resT := produce.ResponseTopic{
					Name: t.Name,
				}
				if err != nil {
					resT.ErrorCode = -1
					log.Errorf("unable to add new kafka record to topic %q: %v", t.Name, err.Error())
				}
				res.Topics = append(res.Topics, resT)
			}
			protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, res)
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
			errCode := protocol.UnknownTopicOrPartition
			if validateTopicName(reqT.Name) != nil {
				errCode = protocol.InvalidTopic
			}
			r.Topics = append(r.Topics, metaData.ResponseTopic{
				ErrorCode: errCode,
				Name:      reqT.Name,
			})
		}
	}

	return r

}

func (s *Binding) processFindCoordinator(req *findCoordinator.Request) *findCoordinator.Response {
	r := &findCoordinator.Response{}

	switch req.KeyType {
	case findCoordinator.KeyTypeGroup:
		g := s.getOrCreateGroup(req.Key)
		r.NodeId = int32(g.coordinator.id)
		r.Host = g.coordinator.host
		r.Port = int32(g.coordinator.port)
	default:
		msg := fmt.Sprintf("unsupported key type '%v' in find coordinator request", req.KeyType)
		log.Error(msg)
		r.ErrorCode = -1
		r.ErrorMessage = msg
	}

	return r
}

func (s *Binding) processJoinGroup(h *protocol.Header, req *joinGroup.Request, consumer *client, w io.Writer) protocol.ErrorCode {
	var g *group
	var exists bool

	if g, exists = s.getGroup(req.GroupId); !exists {
		return protocol.InvalidGroupId
	} else if g.state == awaitSync {
		return protocol.RebalanceInProgress
	} else if g.state == stable {
		g.state = joining
		go g.balancer.startJoin()
	}

	if len(req.MemberId) == 0 {
		memberId := fmt.Sprintf("%v-%v", h.ClientId, utils.NewGuid())
		protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, &joinGroup.Response{
			ErrorCode: protocol.MemberIdRequired,
			MemberId:  memberId,
		})
		return protocol.None
	}

	j := join{
		consumer:  consumer,
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

func (s *Binding) handleSyncGroup(h *protocol.Header, req *syncGroup.Request, consumer *client, w io.Writer) protocol.ErrorCode {
	g, exists := s.getGroup(req.GroupId)
	if !exists {
		return protocol.InvalidGroupId
	}

	switch g.state {
	case joining:
		return protocol.RebalanceInProgress
	case stable:
		if !isMember(consumer, g.members) {
			return protocol.UnknownMemberId
		}
	}

	sync := syncData{
		consumer:     consumer,
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

	return protocol.None
}

func (s *Binding) processOffSetFetch(req *offsetFetch.Request) *offsetFetch.Response {
	r := &offsetFetch.Response{
		Topics: make([]offsetFetch.ResponseTopic, 0, len(req.Topics)),
	}

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

type partitionData struct {
	fetchOffset int64
	set         protocol.RecordBatch
	size        int32
	maxBytes    int32
	error       protocol.ErrorCode
}

type topicData struct {
	partitions map[int32]*partitionData
}

func (s *Binding) processFetch(req *fetch.Request) *fetch.Response {
	r := &fetch.Response{Topics: make([]fetch.ResponseTopic, 0)}

	start := time.Now().Add(time.Duration(req.MaxWaitMs-200) * time.Millisecond) // -200: working load time
	size := int32(0)

	topics := make(map[string]topicData)
	for _, t := range req.Topics {
		topics[t.Name] = topicData{partitions: make(map[int32]*partitionData)}
		for _, p := range t.Partitions {
			topics[t.Name].partitions[p.Index] = &partitionData{fetchOffset: p.FetchOffset, maxBytes: p.MaxBytes}
		}
	}

	for {
		for name, topic := range topics {
			t := s.topics[name]
			for index, partition := range topic.partitions {
				p := t.partitions[int(index)]
				if p.offset > 0 && partition.fetchOffset > p.offset {
					partition.error = protocol.OffsetOutOfRange
				}
				set, offset, setSize := p.read(partition.fetchOffset, partition.maxBytes-partition.size)
				partition.set = set
				partition.fetchOffset = offset
				partition.size += setSize
			}
		}

		if time.Now().After(start) || size > req.MinBytes {
			break
		}

		time.Sleep(time.Duration(math.Floor(0.2*float64(req.MaxWaitMs))) * time.Millisecond)
	}

	for name, topic := range topics {
		t := s.topics[name]
		resTopic := fetch.ResponseTopic{Name: name, Partitions: make([]fetch.ResponsePartition, 0, len(topic.partitions))}
		for index, partition := range topic.partitions {
			p := t.partitions[int(index)]
			resPar := fetch.ResponsePartition{
				Index:                index,
				HighWatermark:        p.offset,
				LastStableOffset:     p.offset,
				LogStartOffset:       p.startOffset,
				PreferredReadReplica: -1,
				RecordSet:            partition.set,
				ErrorCode:            int16(partition.error),
			}
			resTopic.Partitions = append(resTopic.Partitions, resPar)
		}
		r.Topics = append(r.Topics, resTopic)
	}

	return r
}

func (s *Binding) processOffsetCommit(req *offsetCommit.Request) *offsetCommit.Response {
	r := &offsetCommit.Response{
		Topics: make([]offsetCommit.ResponseTopic, 0, len(req.Topics)),
	}

	for _, rt := range req.Topics {
		t := s.topics[rt.Name]
		resTopic := offsetCommit.ResponseTopic{Name: rt.Name, Partitions: make([]offsetCommit.ResponsePartition, 0, len(rt.Partitions))}
		for _, rp := range rt.Partitions {
			p := t.partitions[int(rp.Index)]
			errCode := int16(0)
			if rp.Offset > p.offset {
				errCode = int16(protocol.OffsetOutOfRange)
			} else {
				p.setOffset(req.GroupId, rp.Offset)
			}
			resTopic.Partitions = append(resTopic.Partitions, offsetCommit.ResponsePartition{
				Index:     rp.Index,
				ErrorCode: errCode,
			})
		}
		r.Topics = append(r.Topics, resTopic)
	}

	return r
}
