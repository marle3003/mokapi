package store

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/kafka/findCoordinator"
	"net"
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

	switch r.KeyType {
	case findCoordinator.KeyTypeGroup:
		b := s.getBrokerByHost(req.Host)
		if b == nil {
			return writeError(kafka.UnknownServerError, fmt.Sprintf("broker %v not found", req.Host))
		}
		g := s.GetOrCreateGroup(r.Key, b.Id)
		if g.Coordinator == nil {
			return writeError(kafka.CoordinatorNotAvailable, fmt.Sprintf("no coordinator for group %v available", r.Key))
		} else {
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
			res.Port = int32(b.Port)
		}
	default:
		res.ErrorCode = kafka.UnknownServerError
		res.ErrorMessage = fmt.Sprintf("unsupported request key_type=%v", r.KeyType)
		log.Errorf("kafka FindCoordinator: %v", res.ErrorMessage)
	}

	return rw.Write(res)
}
