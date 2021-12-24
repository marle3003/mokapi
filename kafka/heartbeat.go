package kafka

import (
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/heartbeat"
	"mokapi/kafka/store"
)

func (b *Broker) heartbeat(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*heartbeat.Request)

	ctx := getClientContext(req)
	if _, ok := ctx.member[r.GroupId]; !ok {
		return rw.Write(&heartbeat.Response{ErrorCode: protocol.UnknownMemberId})
	} else {
		g, ok := b.Store.Group(r.GroupId)
		if !ok {
			return rw.Write(&heartbeat.Response{ErrorCode: protocol.InvalidGroupId})
		}
		if g.State() != store.Stable {
			return rw.Write(&heartbeat.Response{ErrorCode: protocol.RebalanceInProgress})
		}
	}

	return rw.Write(&heartbeat.Response{})
}
