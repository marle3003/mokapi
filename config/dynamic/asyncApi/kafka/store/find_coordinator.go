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
			res.ErrorMessage = fmt.Sprintf("broker %v not found", req.Host)
			log.Errorf("kafka FindCoordinator: %v", res.ErrorMessage)
			break
		}
		g := s.GetOrCreateGroup(r.Key, b.Id)
		if g.Coordinator == nil {
			res.ErrorCode = kafka.CoordinatorNotAvailable
			res.ErrorMessage = fmt.Sprintf("no coordinator for group %v available", r.Key)
			log.Errorf("kafka FindCoordinator: %v", res.ErrorMessage)
		} else {
			res.NodeId = int32(b.Id)
			res.Host = b.Host
			res.Port = int32(b.Port)
		}
	default:
		res.ErrorCode = kafka.UnknownServerError
		res.ErrorMessage = fmt.Sprintf("unsupported request key_type=%v", r.KeyType)
		log.Errorf("kafka FindCoordinator: %v", res.ErrorMessage)
	}

	return rw.Write(res)
}
