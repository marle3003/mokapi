package store

import (
	"mokapi/kafka"
	"mokapi/kafka/listgroup"
)

func (s *Store) listgroup(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*listgroup.Request)

	res := &listgroup.Response{}
	for _, g := range s.Groups() {
		group := listgroup.Group{
			GroupId:    g.Name,
			GroupState: "Empty",
		}

		if g.Generation != nil && len(g.Generation.Members) > 0 {
			group.ProtocolType = g.Generation.Protocol
			switch g.State {
			case Joining:
				group.GroupState = "PreparingRebalance"
			case AwaitingSync:
				group.GroupState = "CompletingRebalance"
			case Stable:
				group.GroupState = "Stable"
			}
		}

		if containsState(r.StatesFilter, group.GroupState) {
			res.Groups = append(res.Groups, group)
		}
	}

	return rw.Write(res)
}

func containsState(states []string, state string) bool {
	if len(states) == 0 {
		return true
	}
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}
