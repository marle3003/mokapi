package kafka

import (
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/heartbeat"
)

func (b *BrokerServer) heartbeat(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*heartbeat.Request)

	ctx := getClientContext(req)
	if _, ok := ctx.member[r.GroupId]; !ok {
		return rw.Write(&heartbeat.Response{ErrorCode: protocol.UnknownMemberId})
	} else {
		g := b.Cluster.Group(r.GroupId)
		if g.State() != Stable {
			return rw.Write(&heartbeat.Response{ErrorCode: protocol.RebalanceInProgress})
		}
	}

	return rw.Write(&heartbeat.Response{})
}
