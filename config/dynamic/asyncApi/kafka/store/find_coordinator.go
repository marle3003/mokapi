package store

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/kafka/findCoordinator"
)

func (s *Store) findCoordinator(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*findCoordinator.Request)
	res := &findCoordinator.Response{}

	switch r.KeyType {
	case findCoordinator.KeyTypeGroup:
		b := s.getBrokerByHost(req.Host)
		if b == nil {
			res.ErrorCode = kafka.UnknownServerError
			res.ErrorMessage = "broker not found"
			break
		}
		g := s.GetOrCreateGroup(r.Key, b.Id)

		if g.Coordinator == nil {
			res.ErrorCode = kafka.CoordinatorNotAvailable
		} else {
			res.NodeId = int32(b.Id)
			res.Host = b.Host
			res.Port = int32(b.Port)
		}

	default:
		res.ErrorCode = kafka.UnknownServerError
		res.ErrorMessage = fmt.Sprintf("unsupported key type %v in find coordinator request", r.KeyType)
		log.Errorf(res.ErrorMessage)
	}

	return rw.Write(res)
}
