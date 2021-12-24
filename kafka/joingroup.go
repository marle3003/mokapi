package kafka

import (
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/joinGroup"
)

func (b *Broker) joingroup(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*joinGroup.Request)
	ctx := getClientContext(req)

	g := b.Store.GetOrCreateGroup(r.GroupId, b.Id)
	if g.Coordinator().Id() != b.Id {
		return rw.Write(&joinGroup.Response{ErrorCode: protocol.NotCoordinator})
	}
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
