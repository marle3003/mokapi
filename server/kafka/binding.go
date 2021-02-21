package kafka

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	event "mokapi/models/eventService"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/apiVersion"
	"mokapi/server/kafka/protocol/findCoordinator"
	"mokapi/server/kafka/protocol/metaData"
	"net"
)

type Binding struct {
	stop      chan bool
	listen    string
	isRunning bool
	service   *event.Service
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
			handleApiVersion(h, msg.(*apiVersion.Request), conn)
		case protocol.Metadata:
			handleMetadata(h, msg.(*metaData.Request), conn)
		case protocol.FindCoordinator:
			_ = msg.(*findCoordinator.Request)
		}
	}
}

func handleApiVersion(h *protocol.Header, msg *apiVersion.Request, w io.Writer) {
	apiKeys := make([]apiVersion.ApiKeyResponse, len(protocol.ApiTypes))
	i := 0
	for k, t := range protocol.ApiTypes {
		apiKeys[i] = apiVersion.ApiKeyResponse{ApiKey: k, MinVersion: t.MinVersion, MaxVersion: t.MaxVersion}
		i++
	}

	res := &apiVersion.Response{
		ErrorCode:      0,
		ApiKeys:        apiKeys,
		ThrottleTimeMs: 0,
		TagFields:      nil,
	}

	protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, res)
}

func handleMetadata(h *protocol.Header, msg *metaData.Request, w io.Writer) {
	res := &metaData.Response{
		ThrottleTimeMs: 0,
		Brokers: []metaData.ResponseBroker{
			{
				NodeId:    0,
				Host:      "localhost",
				Port:      9092,
				Rack:      "",
				TagFields: nil,
			},
		},
		ClusterId:    "",
		ControllerId: 0,
		Topics: []metaData.ResponseTopic{
			{
				ErrorCode:  0,
				Name:       "message",
				IsInternal: false,
				Partitions: []metaData.ResponsePartition{
					{
						ErrorCode:       0,
						PartitionIndex:  0,
						LeaderId:        0,
						LeaderEpoch:     0,
						ReplicaNodes:    make([]int32, 0),
						IsrNodes:        make([]int32, 0),
						OfflineReplicas: make([]int32, 0),
						TagFields:       nil,
					},
				},
				TopicAuthorizedOperations: 0,
				TagFields:                 nil,
			},
		},
		ClusterAuthorizedOperations: 0,
	}

	protocol.WriteMessage(w, h.ApiKey, h.ApiVersion, h.CorrelationId, res)
}
