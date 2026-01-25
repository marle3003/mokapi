package store

import (
	"mokapi/kafka"
	"mokapi/kafka/heartbeat"

	log "github.com/sirupsen/logrus"
)

func (s *Store) heartbeat(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*heartbeat.Request)

	ctx := kafka.ClientFromContext(req.Context)
	if _, ok := ctx.Member[r.GroupId]; !ok {
		log.Errorf("kafka Heartbeat: unknown member %v", ctx.ClientId)
		return rw.Write(&heartbeat.Response{ErrorCode: kafka.UnknownMemberId})
	} else {
		g, ok := s.Group(r.GroupId)
		if !ok {
			log.Errorf("kafka Heartbeat: invalid group %v", r.GroupId)
			return rw.Write(&heartbeat.Response{ErrorCode: kafka.InvalidGroupId})
		}
		if g.State != Stable {
			return rw.Write(&heartbeat.Response{ErrorCode: kafka.RebalanceInProgress})
		}
	}

	return rw.Write(&heartbeat.Response{})
}
