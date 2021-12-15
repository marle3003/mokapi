package kafka

import (
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/syncGroup"
)

func (b *BrokerServer) syncgroup(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*syncGroup.Request)
	ctx := getClientContext(req)

	if len(r.MemberId) == 0 {
		return rw.Write(&syncGroup.Response{ErrorCode: protocol.MemberIdRequired})
	}

	g := b.Cluster.Group(r.GroupId)
	balancer := b.getBalancer(g)

	if g.State() == Joining {
		return rw.Write(&syncGroup.Response{ErrorCode: protocol.RebalanceInProgress})
	}

	gen := g.Generation()
	if gen == nil || gen.Id != int(r.GenerationId) {
		return rw.Write(&syncGroup.Response{ErrorCode: protocol.IllegalGeneration})
	}

	if _, ok := ctx.member[r.GroupId]; !ok {
		return rw.Write(&syncGroup.Response{ErrorCode: protocol.RebalanceInProgress})
	}

	data := syncdata{
		client: ctx,
		writer: rw,
	}

	if len(r.GroupAssignments) > 0 {
		data.assigns = make(map[string]*groupAssignment)
		for _, assign := range r.GroupAssignments {
			data.assigns[assign.MemberId] = newGroupAssignment(assign.Assignment)
		}
	}

	// balancer writes the response
	balancer.sync <- data

	return nil
}
