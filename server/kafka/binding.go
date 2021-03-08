package kafka

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"mokapi/models/event"
	"mokapi/models/event/kafka"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/apiVersion"
	"mokapi/server/kafka/protocol/fetch"
	"mokapi/server/kafka/protocol/findCoordinator"
	"mokapi/server/kafka/protocol/heartbeat"
	"mokapi/server/kafka/protocol/joinGroup"
	"mokapi/server/kafka/protocol/listOffsets"
	"mokapi/server/kafka/protocol/metaData"
	"mokapi/server/kafka/protocol/offsetFetch"
	"mokapi/server/kafka/protocol/produce"
	"mokapi/server/kafka/protocol/syncGroup"
	"net"
)

type Binding struct {
	stop      chan bool
	listen    string
	isRunning bool
	service   *event.Service
	brokers   []broker
	groups    map[string]*group
	topics    map[string]*topic
	config    kafka.Binding
}

func NewServer(addr string, c kafka.Binding) *Binding {
	s := &Binding{
		stop:   make(chan bool),
		listen: addr,
		groups: make(map[string]*group),
		topics: make(map[string]*topic),
		config: c,
	}

	b := newBroker(1, "localhost", 9092) // id is 1 based
	s.brokers = append(s.brokers, b)

	return s
}

func (s *Binding) Apply(data interface{}) error {
	service, ok := data.(*event.Service)
	if !ok {
		return errors.Errorf("unexpected parameter type %T in kafka binding", data)
	}
	s.service = service

	for n, c := range service.Channels {
		name := n[1:] // remove leading slash from name
		if _, ok := s.topics[name]; !ok {
			s.topics[name] = &topic{partitions: map[int]*partition{
				0: {leader: s.brokers[0], log: &batchLog{
					batches: make([]*protocol.RecordBatch, 0, 10),
				}}},
			}
			if c.Publish != nil && c.Publish.Message != nil {
				go producer(s.topics[name], c.Publish.Message.ContentType, c.Publish.Message.Payload, s.stop)
			}
		}
	}

	shouldRestart := false
	//if s.listen != "" && s.listen != config.Address {
	//	s.stop <- true
	//	shouldRestart = true
	//}
	//
	//s.listen = config.Address
	//s.listen = "0.0.0.0:9092"

	if s.isRunning {
		log.Infof("Updated configuration of ldap server: %v", s.listen)

		if shouldRestart {
			go s.Start()
		}
	}
	return nil
}

func (s *Binding) Stop() {
	s.stop <- true
}

func (s *Binding) Start() {
	s.isRunning = true

	l, err := net.Listen("tcp", s.listen)
	if err != nil {
		log.Errorf("Error listening: %v", err.Error())
		return
	}

	log.Infof("Started kafka server on %v", s.listen)

	// Close the listener when the application closes.
	connChannl := make(chan net.Conn)
	close := make(chan bool)
	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-close:
					return
				default:
					log.Errorf("Error accepting: %v", err.Error())
				}
			}
			// Handle connections in a new goroutine.
			connChannl <- conn
		}
	}()

	go func() {
		for {
			select {
			case conn := <-connChannl:
				go s.handle(conn)
			case <-s.stop:
				log.Infof("Stopping ldap server on %v", s.listen)
				close <- true
				l.Close()
			}
		}
	}()
}

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

		func() {
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
				r := s.processFetch(msg.(*fetch.Request))
				protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, r)
			case protocol.Heartbeat:
				protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, &heartbeat.Response{})
			case protocol.Produce:
				_ = msg.(*produce.Request)
			case protocol.ListOffsets:
				r := s.processListOffsets(msg.(*listOffsets.Request))
				protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, r)
			}
		}()
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
					Timestamp: 0,
				}

				if rp.Timestamp == -2 { // latest
					if len(p.log.batches) == 0 {
						part.Offset = -1
					} else {
						part.Offset = 0
					}

				} else if rp.Timestamp == -1 { // earliest
					part.Offset = p.log.offset - 1
				}

				if part.Offset >= 0 {
					part.Timestamp = protocol.Timestamp(p.log.batches[0].Records[0].Time)
				}

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
			Host:   "localhost",
			Port:   9092,
		})
	}

	r.ControllerId = r.Brokers[0].NodeId // using first broker as controller

	if len(req.Topics) == 0 {
		for n, t := range s.topics {
			resT := metaData.ResponseTopic{
				Name:       n,
				Partitions: make([]metaData.ResponsePartition, 0, len(t.partitions)),
			}

			for _, p := range t.partitions {
				resT.Partitions = append(resT.Partitions, metaData.ResponsePartition{
					PartitionIndex: 0,
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

			for _, p := range t.partitions {
				resT.Partitions = append(resT.Partitions, metaData.ResponsePartition{
					PartitionIndex: 0,
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
				coordinator: s.brokers[0],
			}
			g.balancer = newGroupBalancer(g, s.config.Group.RebalanceDelay)
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

func (s *Binding) processJoinGroup(h *protocol.Header, req *joinGroup.Request, w io.Writer) int16 {
	var g *group
	var exists bool
	if g, exists = s.groups[req.GroupId]; !exists {
		return -1
	} else if g.state == completingRebalance {
		return 27 // RebalanceInProgressCode
	} else if g.state == empty || g.state == stable {
		g.state = preparingRebalance
		go g.balancer.startJoin()
	}

	consumer := consumer{id: req.MemberId}

	if len(consumer.id) == 0 {
		consumer.id = fmt.Sprintf("%v-%v", h.ClientId, createGuid())
	}

	j := join{
		consumer:  consumer,
		protocols: make([]groupAssignmentStrategy, 0, len(req.Protocols)),
		write: func(msg protocol.Message) {
			protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, msg)
		},
		rebalanceTimeout: int(req.RebalanceTimeoutMs),
	}

	for _, p := range req.Protocols {
		j.protocols = append(j.protocols, groupAssignmentStrategy{
			assignmentStrategy: p.Name,
			metadata:           p.MetaData,
		})
	}

	g.balancer.join <- j

	return 0
}

func (s *Binding) handleSyncGroup(h *protocol.Header, req *syncGroup.Request, w io.Writer) int {
	var g *group
	var exists bool
	if g, exists = s.groups[req.GroupId]; !exists {
		return -1
	}

	sync := sync{
		consumer:     consumer{id: req.MemberId},
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
				CommittedOffset: p.log.committed,
			})
		}
		r.Topics = append(r.Topics, resTopic)
	}

	return r
}

func (s *Binding) processFetch(req *fetch.Request) *fetch.Response {
	r := &fetch.Response{Topics: make([]fetch.ResponseTopic, 0)}

	// currently offset is not separated by groups
	for _, rt := range req.Topics {
		t := s.topics[rt.Name]
		resTopic := fetch.ResponseTopic{Name: rt.Name, Partitions: make([]fetch.ResponsePartition, 0, len(rt.Partitions))}
		for _, rp := range rt.Partitions {
			p := t.partitions[int(rp.Index)]
			resPar := fetch.ResponsePartition{
				Index:                rp.Index,
				HighWatermark:        p.log.offset - 1,
				LastStableOffset:     p.log.offset - 1,
				LogStartOffset:       0,
				PreferredReadReplica: 1,
			}

			//set := protocol.RecordSet{Batches: make([]protocol.RecordBatch, 0)}
			//if len(p.log.batches) > int(rp.FetchOffset) {
			//	batch := p.log.batches[rp.FetchOffset]
			//
			//	set.Batches = append(set.Batches, *batch)
			//}

			size := int32(0)
			set := protocol.RecordSet{Batches: make([]protocol.RecordBatch, 0)}
			for _, b := range p.log.batches[rp.FetchOffset:] {
				set.Batches = append(set.Batches, *b)
				size += b.Size()
				if size > rp.MaxBytes {
					break
				}
			}
			resPar.RecordSet = set
			resTopic.Partitions = append(resTopic.Partitions, resPar)
		}
		r.Topics = append(r.Topics, resTopic)
	}

	return r
}
