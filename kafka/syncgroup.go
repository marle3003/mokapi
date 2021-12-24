package kafka

import (
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/syncGroup"
	"mokapi/kafka/store"
)

func (b *Broker) syncgroup(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*syncGroup.Request)
	ctx := getClientContext(req)

	if len(r.MemberId) == 0 {
		return rw.Write(&syncGroup.Response{ErrorCode: protocol.MemberIdRequired})
	}

	g := b.Store.GetOrCreateGroup(r.GroupId, b.Id)
	if g.Coordinator().Id() != b.Id {
		return rw.Write(&syncGroup.Response{ErrorCode: protocol.NotCoordinator})
	}
	balancer := b.getBalancer(g)

	if g.State() == store.Joining {
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
