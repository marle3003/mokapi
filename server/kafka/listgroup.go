package kafka

import (
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/listgroup"
)

func (b *BrokerServer) listgroup(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*listgroup.Request)

	res := &listgroup.Response{}
	for _, g := range b.Cluster.Groups() {
		group := listgroup.Group{
			GroupId:    g.Name(),
			GroupState: "Empty",
		}

		gen := g.Generation()
		if gen != nil && len(gen.Members) > 0 {
			group.ProtocolType = gen.Protocol
			switch g.State() {
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
