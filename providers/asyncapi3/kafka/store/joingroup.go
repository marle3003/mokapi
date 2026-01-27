package store

import (
	"mokapi/kafka"
	"mokapi/kafka/joinGroup"
	"mokapi/runtime/events"
)

func (s *Store) joingroup(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*joinGroup.Request)
	ctx := kafka.ClientFromContext(req.Context)

	b := s.getBrokerByHost(req.Host)
	if b == nil {
		res := &joinGroup.Response{
			ErrorCode: kafka.UnknownServerError,
		}
		return rw.Write(res)
	}

	g := s.GetOrCreateGroup(r.GroupId, b.Id)

	ctx.AddGroup(g.Name, r.MemberId)

	data := joindata{
		client:           ctx,
		writer:           rw,
		protocolType:     r.ProtocolType,
		protocols:        r.Protocols,
		rebalanceTimeout: int(r.RebalanceTimeoutMs),
		sessionTimeout:   int(r.SessionTimeoutMs),
		log:              s.logJoinGroupRequest,
	}

	// balancer writes the response
	g.balancer.join <- data

	return nil
}

func (s *Store) logJoinGroupRequest(log *KafkaRequestLog, clientId string) {
	log.Api = s.cluster
	t := events.NewTraits().
		WithNamespace("kafka").
		WithName(s.cluster).
		With("type", "request").
		With("clientId", clientId)
	_ = s.eh.Push(log, t)
}
