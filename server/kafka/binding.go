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
}

func NewServer(addr string) *Binding {
	s := &Binding{stop: make(chan bool), listen: addr}
	return s
}

func (s *Binding) Apply(data interface{}) error {
	service, ok := data.(*event.Service)
	if !ok {
		return errors.Errorf("unexpected parameter type %T in kafka binding", data)
	}
	s.service = service

	b := broker{Id: 0}
	s.brokers = append(s.brokers, b)

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

		switch h.ApiKey {
		case protocol.ApiVersions:
			s.handleApiVersion(h, msg.(*apiVersion.Request), conn)
		case protocol.Metadata:
			s.handleMetadata(h, msg.(*metaData.Request), conn)
		case protocol.FindCoordinator:
			s.handleFindCoordinator(h, msg.(*findCoordinator.Request), conn)
		case protocol.JoinGroup:
			s.handleJoinGroup(h, msg.(*joinGroup.Request), conn)
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
	}
}

func (s *Binding) handleApiVersion(h *protocol.Header, req *apiVersion.Request, w io.Writer) {
	apiKeys := make([]apiVersion.ApiKeyResponse, len(protocol.ApiTypes))
	i := 0
	for k, t := range protocol.ApiTypes {
		apiKeys[i] = apiVersion.ApiKeyResponse{ApiKey: k, MinVersion: t.MinVersion, MaxVersion: t.MaxVersion}
		i++
	}

	res := &apiVersion.Response{ApiKeys: apiKeys}

	protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, res)
}

func (s *Binding) handleMetadata(h *protocol.Header, req *metaData.Request, w io.Writer) {
	var brokers []metaData.ResponseBroker
	for _, b := range s.brokers {
		brokers = append(brokers, metaData.ResponseBroker{
			NodeId: int32(b.Id),
			Host:   "localhost",
			Port:   9092,
		})
	}
	res := &metaData.Response{
		ThrottleTimeMs:              0,
		Brokers:                     brokers,
		Topics:                      make([]metaData.ResponseTopic, 0),
		ClusterId:                   "mokapi",
		ControllerId:                1,
		ClusterAuthorizedOperations: 0,
	}

	for _, t := range req.Topics {
		name := fmt.Sprintf("/%v", t.Name)
		if _, ok := s.service.Channels[name]; ok {
			res.Topics = append(res.Topics, metaData.ResponseTopic{
				ErrorCode:  0,
				Name:       t.Name,
				IsInternal: false,
				Partitions: []metaData.ResponsePartition{
					{
						ErrorCode:       0,
						PartitionIndex:  0,
						LeaderId:        1,
						LeaderEpoch:     0,
						ReplicaNodes:    []int32{1},
						IsrNodes:        []int32{1},
						OfflineReplicas: make([]int32, 0),
						TagFields:       nil,
					},
				},
				TopicAuthorizedOperations: 0,
				TagFields:                 nil,
			})
		} else {
			res.Topics = append(res.Topics, metaData.ResponseTopic{
				ErrorCode: protocol.UnknownTopicOrPartition,
				Name:      t.Name,
			})
		}
	}

	protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, res)
}

func (s *Binding) handleFindCoordinator(h *protocol.Header, req *findCoordinator.Request, w io.Writer) {
	res := &findCoordinator.Response{
		ThrottleTimeMs: 0,
		ErrorCode:      0,
		ErrorMessage:   "",
		NodeId:         1,
		Host:           "localhost",
		Port:           9092,
		TagFields:      nil,
	}

	protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, res)
}

func (s *Binding) handleJoinGroup(h *protocol.Header, req *joinGroup.Request, w io.Writer) {
	res := &joinGroup.Response{
		ThrottleTimeMs: 0,
		ErrorCode:      0,
		GenerationId:   -1,
		ProtocolName:   "range",
		Leader:         "1",
		MemberId:       "1",
		Members: []joinGroup.Member{
			{
				MemberId:        "1",
				GroupInstanceId: "",
				MetaData:        req.Protocols[0].MetaData,
			},
		},
	}

	protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, res)
}

func (s *Binding) handleSyncGroup(h *protocol.Header, req *syncGroup.Request, w io.Writer) {
	res := &syncGroup.Response{
		ThrottleTimeMs: 0,
		ErrorCode:      0,
		Assignment:     req.GroupAssignments[0].Assignment,
	}

	protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, res)
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
