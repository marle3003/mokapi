package store

import (
	"mokapi/kafka"
	"mokapi/kafka/initProducerId"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

func (s *Store) initProducerID(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*initProducerId.Request)
	res := &initProducerId.Response{}

	if r.TransactionalId == "" {
		if r.ProducerId > 0 {
			ps, ok := s.producers[r.ProducerId]
			if !ok {
				res.ErrorCode = kafka.UnknownProducerId
			} else if r.ProducerEpoch < ps.ProducerEpoch {
				res.ErrorCode = kafka.ProducerFenced
			} else {
				ps.ProducerEpoch++
				res.ProducerId = ps.ProducerId
				res.ProducerEpoch = ps.ProducerEpoch
			}
			return rw.Write(res)
		}

		res.ProducerId = atomic.AddInt64(&s.nextPID, 1)
		res.ProducerEpoch = 0
		ps := &ProducerState{ProducerId: res.ProducerId, ProducerEpoch: res.ProducerEpoch}
		s.producers[res.ProducerId] = ps
	} else {
		res.ErrorCode = kafka.UnsupportedForMessageFormat
		log.Errorf("kafka: mokapi does not support transactional producer: %s", r.TransactionalId)
	}

	return rw.Write(res)
}
