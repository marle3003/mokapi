package store

import (
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/kafka/heartbeat"
)

func (s *Store) heartbeat(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*heartbeat.Request)

	ctx := kafka.ClientFromContext(req)
	if _, ok := ctx.Member[r.GroupId]; !ok {
		log.Errorf("kafka: heartbeat unknown member %v", ctx.ClientId)
		return rw.Write(&heartbeat.Response{ErrorCode: kafka.UnknownMemberId})
	} else {
		g, ok := s.Group(r.GroupId)
		if !ok {
			log.Errorf("kafka: heartbeat invalid group %v", r.GroupId)
			return rw.Write(&heartbeat.Response{ErrorCode: kafka.InvalidGroupId})
		}
		if g.State != Stable {
			return rw.Write(&heartbeat.Response{ErrorCode: kafka.RebalanceInProgress})
		}
	}

	return rw.Write(&heartbeat.Response{})
}
