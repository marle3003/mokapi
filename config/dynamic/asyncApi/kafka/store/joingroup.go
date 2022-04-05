package store

import (
	"mokapi/kafka"
	"mokapi/kafka/joinGroup"
)

func (s *Store) joingroup(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*joinGroup.Request)
	ctx := kafka.ClientFromContext(req)

	b := s.getBrokerByHost(req.Host)
	if b == nil {
		res := &joinGroup.Response{
			ErrorCode: kafka.UnknownServerError,
		}
		return rw.Write(res)
	}

	g := s.GetOrCreateGroup(r.GroupId, b.Id)
	if g.Coordinator.Id != b.Id {
		return rw.Write(&joinGroup.Response{ErrorCode: kafka.NotCoordinator})
	}

	ctx.AddGroup(g.Name, r.MemberId)

	data := joindata{
		client:           ctx,
		writer:           rw,
		protocols:        r.Protocols,
		rebalanceTimeout: int(r.RebalanceTimeoutMs),
		sessionTimeout:   int(r.SessionTimeoutMs),
	}

	// balancer writes the response
	g.balancer.join <- data

	return nil
}
