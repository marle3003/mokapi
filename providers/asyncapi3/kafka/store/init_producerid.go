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

		} else {
			res.ProducerId = atomic.AddInt64(&s.nextPID, 1)
			res.ProducerEpoch = 0
			ps := &ProducerState{ProducerId: res.ProducerId, ProducerEpoch: res.ProducerEpoch}
			s.producers[res.ProducerId] = ps
		}
	} else {
		res.ErrorCode = kafka.UnsupportedForMessageFormat
		log.Errorf("kafka: mokapi does not support transactional producer: %s", r.TransactionalId)
	}

	go func() {
		s.logRequest(req.Header, newKafkaInitProducerIdRequest(r))(newKafkaInitProducerIdResponse(res))
	}()

	return rw.Write(res)
}

func newKafkaInitProducerIdRequest(req *initProducerId.Request) *KafkaInitProducerIdRequest {
	return &KafkaInitProducerIdRequest{
		TransactionalId:      req.TransactionalId,
		TransactionTimeoutMs: req.TransactionTimeoutMs,
		ProducerId:           req.ProducerId,
		ProducerEpoch:        req.ProducerEpoch,
		Enable2PC:            req.Enable2PC,
	}
}

func newKafkaInitProducerIdResponse(res *initProducerId.Response) *KafkaInitProducerIdResponse {
	r := &KafkaInitProducerIdResponse{
		ProducerId:              res.ProducerId,
		ProducerEpoch:           res.ProducerEpoch,
		OngoingTxnProducerId:    res.OngoingTxnProducerId,
		OngoingTxnProducerEpoch: res.OngoingTxnProducerEpoch,
	}
	if res.ErrorCode != kafka.None {
		r.ErrorCode = res.ErrorCode.String()
	}
	return r
}
