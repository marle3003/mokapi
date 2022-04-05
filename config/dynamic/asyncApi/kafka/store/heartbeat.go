package store

import (
	"mokapi/kafka"
	"mokapi/kafka/heartbeat"
)

func (s *Store) heartbeat(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*heartbeat.Request)

	ctx := kafka.ClientFromContext(req)
	if _, ok := ctx.Member[r.GroupId]; !ok {
		return rw.Write(&heartbeat.Response{ErrorCode: kafka.UnknownMemberId})
	} else {
		g, ok := s.Group(r.GroupId)
		if !ok {
			return rw.Write(&heartbeat.Response{ErrorCode: kafka.InvalidGroupId})
		}
		if g.State != Stable {
			return rw.Write(&heartbeat.Response{ErrorCode: kafka.RebalanceInProgress})
		}
	}

	return rw.Write(&heartbeat.Response{})
}
