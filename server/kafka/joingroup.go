package kafka

import (
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/joinGroup"
)

func (b *BrokerServer) joingroup(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*joinGroup.Request)
	ctx := getClientContext(req)

	if len(r.MemberId) == 0 {
		return rw.Write(&joinGroup.Response{ErrorCode: protocol.MemberIdRequired})
	}

	g := b.Cluster.Group(r.GroupId)
	balancer := b.getBalancer(g)

	ctx.AddGroup(g.Name(), r.MemberId)

	data := joindata{
		client:           ctx,
		writer:           rw,
		protocols:        r.Protocols,
		rebalanceTimeout: int(r.RebalanceTimeoutMs),
		sessionTimeout:   int(r.SessionTimeoutMs),
	}

	// balancer writes the response
	balancer.join <- data

	return nil
}
