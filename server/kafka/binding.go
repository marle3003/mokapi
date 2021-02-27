package kafka

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	event "mokapi/models/eventService"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/apiVersion"
	"mokapi/server/kafka/protocol/fetch"
	"mokapi/server/kafka/protocol/findCoordinator"
	"mokapi/server/kafka/protocol/heartbeat"
	"mokapi/server/kafka/protocol/joinGroup"
	"mokapi/server/kafka/protocol/metaData"
	"mokapi/server/kafka/protocol/offsetFetch"
	"mokapi/server/kafka/protocol/produce"
	"mokapi/server/kafka/protocol/syncGroup"
	"net"
	"time"
)

type Binding struct {
	stop      chan bool
	listen    string
	isRunning bool
	service   *event.Service
	brokers   []broker
	groups    map[string]*group
	topics    map[string]topic
}

func NewServer(addr string) *Binding {
	s := &Binding{
		stop:   make(chan bool),
		listen: addr,
		groups: make(map[string]*group),
		topics: make(map[string]topic),
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

	for n, _ := range service.Channels {
		name := n[1:] // remove leading slash from name
		s.topics[name] = topic{partitions: map[int]partition{
			0: {leader: s.brokers[0]}},
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

		go func() {
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
				s.handleOffSetFetch(h, msg.(*offsetFetch.Request), conn)
			case protocol.Fetch:
				s.handleFetch(h, msg.(*fetch.Request), conn)
			case protocol.Heartbeat:
				protocol.WriteMessage(conn, h.ApiKey, h.ApiVersion, h.CorrelationId, &heartbeat.Response{})
			case protocol.Produce:
				_ = msg.(*produce.Request)
			}
		}()
	}
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
			g.balancer = newGroupBalancer(g)
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
	} else if g.state == syncing {
		return 27 // RebalanceInProgressCode
	} else if g.state == empty {
		g.state = joining
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

func (s *Binding) handleOffSetFetch(h *protocol.Header, req *offsetFetch.Request, w io.Writer) {
	res := &offsetFetch.Response{
		ThrottleTimeMs: 0,
		Topics: []offsetFetch.ResponseTopic{
			{
				Name: "message",
				Partitions: []offsetFetch.Partition{
					{
						Index:           0,
						CommittedOffset: 0,
						Metadata:        "",
						ErrorCode:       0,
					},
				},
			},
		},
		ErrorCode: 0,
	}

	protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, res)
}

func (s *Binding) handleFetch(h *protocol.Header, req *fetch.Request, w io.Writer) {
	time.Sleep(time.Millisecond * time.Duration(req.MaxWaitMs-50))
	res := &fetch.Response{
		Topics: []fetch.ResponseTopic{
			{
				Name: "message",
				Partitions: []fetch.ResponsePartition{
					{
						Index:                0,
						ErrorCode:            0,
						HighWatermark:        1,
						LastStableOffset:     1,
						PreferredReadReplica: -1,
						RecordSet: protocol.RecordSet{
							Batches: []protocol.RecordBatch{
								{
									Attributes: 0,
									ProducerId: 0,
									Records: []protocol.Record{
										{
											Key:     nil,
											Value:   []byte("Test"),
											Headers: nil,
										},
										{
											Key:     nil,
											Value:   []byte("Test2"),
											Headers: nil,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, res)
}
