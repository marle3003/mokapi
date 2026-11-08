package store

import (
	"mokapi/kafka"
	"mokapi/kafka/initProducerId"

	log "github.com/sirupsen/logrus"
)

func (s *Store) initProducerID(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*initProducerId.Request)
	res := &initProducerId.Response{}

	if r.TransactionalId == "" {
		var ps ProducerState
		if r.ProducerId > 0 {
			ps, res.ErrorCode = s.renewProducer(r.ProducerId, r.ProducerEpoch)
		} else {
			ps = s.registerProducer()
		}
		res.ProducerId = ps.ProducerId
		res.ProducerEpoch = ps.ProducerEpoch
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
