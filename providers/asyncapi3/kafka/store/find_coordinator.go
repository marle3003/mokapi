package store

import (
	"fmt"
	"mokapi/kafka"
	"mokapi/kafka/findCoordinator"
	"net"

	log "github.com/sirupsen/logrus"
)

func (s *Store) findCoordinator(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*findCoordinator.Request)
	res := &findCoordinator.Response{}

	writeError := func(code kafka.ErrorCode, msg string) error {
		res.ErrorCode = code
		res.ErrorMessage = msg
		log.Errorf("kafka FindCoordinator: %v", msg)
		return rw.Write(res)
	}

	reqLog := &KafkaFindCoordinatorRequest{
		Key:     r.Key,
		KeyType: r.KeyType,
	}
	resLog := &KafkaFindCoordinatorResponse{}

	switch r.KeyType {
	case findCoordinator.KeyTypeGroup:
		b := s.getBrokerByHost(req.Host)
		if b == nil {
			return writeError(kafka.UnknownServerError, fmt.Sprintf("broker %v not found", req.Host))
		}
		host := b.Host
		if len(host) == 0 {
			var err error
			host, _, err = net.SplitHostPort(req.Host)
			if err != nil {
				return writeError(kafka.UnknownServerError, fmt.Sprintf("broker %v not found: %v", req.Host, err))
			}
		}

		res.NodeId = int32(b.Id)
		res.Host = host
		resLog.Host = host
		res.Port = int32(b.Port)
		resLog.Port = b.Port
	default:
		res.ErrorCode = kafka.UnknownServerError
		res.ErrorMessage = fmt.Sprintf("unsupported request key_type=%v", r.KeyType)
		log.Errorf("kafka FindCoordinator: %v", res.ErrorMessage)
		resLog.ErrorMessage = fmt.Sprintf("unsupported request key_type=%v", r.KeyType)
		resLog.ErrorCode = kafka.UnknownServerError.String()
	}

	go func() {
		l := &KafkaRequestLogEvent{
			Request:  reqLog,
			Response: resLog,
		}
		s.logRequest(req.Header)(l)
	}()

	return rw.Write(res)
}
