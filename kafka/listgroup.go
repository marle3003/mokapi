package kafka

import (
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/listgroup"
	"mokapi/kafka/store"
)

func (b *Broker) listgroup(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*listgroup.Request)

	res := &listgroup.Response{}
	for _, g := range b.Store.Groups() {
		group := listgroup.Group{
			GroupId:    g.Name(),
			GroupState: "Empty",
		}

		gen := g.Generation()
		if gen != nil && len(gen.Members) > 0 {
			group.ProtocolType = gen.Protocol
			switch g.State() {
			case store.Joining:
				group.GroupState = "PreparingRebalance"
			case store.AwaitingSync:
				group.GroupState = "CompletingRebalance"
			case store.Stable:
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
