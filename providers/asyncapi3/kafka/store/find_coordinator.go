package store

import (
	"fmt"
	"mokapi/kafka"
	"mokapi/kafka/findCoordinator"

	log "github.com/sirupsen/logrus"
)

func (s *Store) findCoordinator(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*findCoordinator.Request)
	res := &findCoordinator.Response{}

	reqLog := &KafkaFindCoordinatorRequest{
		Key:     r.Key,
		KeyType: r.KeyType,
	}
	resLog := &KafkaFindCoordinatorResponse{}

	switch r.KeyType {
	case findCoordinator.KeyTypeGroup:
		host, port := parseHostAndPort(req.Host)
		// Mokapi does no leader management: always return fixed node id
		res.NodeId = 0
		res.Host = host
		res.Port = int32(port)
	default:
		res.ErrorCode = kafka.UnknownServerError
		res.ErrorMessage = fmt.Sprintf("unsupported request key_type=%v", r.KeyType)
		log.Errorf("kafka FindCoordinator: %v", res.ErrorMessage)
		resLog.ErrorMessage = fmt.Sprintf("unsupported request key_type=%v", r.KeyType)
		resLog.ErrorCode = kafka.UnknownServerError.String()
	}

	go func() {
		s.logRequest(req.Header, reqLog)(newKafkaFindCoordinatorResponse(res))
	}()

	return rw.Write(res)
}

func newKafkaFindCoordinatorResponse(res *findCoordinator.Response) *KafkaFindCoordinatorResponse {
	r := &KafkaFindCoordinatorResponse{
		Host: res.Host,
		Port: int(res.Port),
	}
	if res.ErrorCode != kafka.None {
		r.ErrorCode = res.ErrorCode.String()
		r.ErrorMessage = res.ErrorMessage
	}
	return r
}
