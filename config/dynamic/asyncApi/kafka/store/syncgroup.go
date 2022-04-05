package store

import (
	"mokapi/kafka"
	"mokapi/kafka/syncGroup"
)

func (s *Store) syncgroup(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*syncGroup.Request)
	ctx := kafka.ClientFromContext(req)

	if len(r.MemberId) == 0 {
		return rw.Write(&syncGroup.Response{ErrorCode: kafka.MemberIdRequired})
	}

	b := s.getBrokerByHost(req.Host)
	if b == nil {
		res := &syncGroup.Response{
			ErrorCode: kafka.UnknownServerError,
		}
		return rw.Write(res)
	}

	g := s.GetOrCreateGroup(r.GroupId, b.Id)
	if g.Coordinator.Id != b.Id {
		return rw.Write(&syncGroup.Response{ErrorCode: kafka.NotCoordinator})
	}

	if g.State == Joining {
		return rw.Write(&syncGroup.Response{ErrorCode: kafka.RebalanceInProgress})
	}

	if g.Generation == nil || g.Generation.Id != int(r.GenerationId) {
		return rw.Write(&syncGroup.Response{ErrorCode: kafka.IllegalGeneration})
	}

	if _, ok := ctx.Member[r.GroupId]; !ok {
		return rw.Write(&syncGroup.Response{ErrorCode: kafka.RebalanceInProgress})
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
	g.balancer.sync <- data

	return nil
}
