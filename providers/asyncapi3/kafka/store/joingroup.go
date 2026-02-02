package store

import (
	"mokapi/kafka"
	"mokapi/kafka/joinGroup"
)

func (s *Store) joingroup(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*joinGroup.Request)
	ctx := kafka.ClientFromContext(req.Context)

	reqLog := newKafkaJoinGroupRequest(r)

	data := joindata{
		client:           ctx,
		writer:           rw,
		protocolType:     r.ProtocolType,
		protocols:        r.Protocols,
		rebalanceTimeout: int(r.RebalanceTimeoutMs),
		sessionTimeout:   int(r.SessionTimeoutMs),
		log:              s.logRequest(req.Header, reqLog),
	}

	b := s.getBrokerByPort(req.Host)
	if b == nil {
		res := &joinGroup.Response{
			ErrorCode: kafka.UnknownServerError,
		}
		go func() {
			resLog := &KafkaJoinGroupResponse{}
			resLog.ErrorCode = res.ErrorCode.String()

			s.logRequest(req.Header, reqLog)(&KafkaRequestLogEvent{
				Response: resLog,
			})
		}()
		return rw.Write(res)
	}

	g := s.GetOrCreateGroup(r.GroupId, b.Id)

	ctx.AddGroup(g.Name, r.MemberId)

	// balancer writes the response
	g.balancer.join <- data

	return nil
}

func newKafkaJoinGroupRequest(req *joinGroup.Request) *KafkaJoinGroupRequest {
	r := &KafkaJoinGroupRequest{
		GroupName:    req.GroupId,
		MemberId:     req.MemberId,
		ProtocolType: req.ProtocolType,
	}
	for _, proto := range req.Protocols {
		r.Protocols = append(r.Protocols, proto.Name)
	}
	return r
}
