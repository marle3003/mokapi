package kafka

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/findCoordinator"
)

func (b *BrokerServer) findCoordinator(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*findCoordinator.Request)
	res := &findCoordinator.Response{}

	switch r.KeyType {
	case findCoordinator.KeyTypeGroup:
		g := b.Cluster.Group(r.Key)
		if g == nil {
			res.ErrorCode = protocol.InvalidGroupId
		} else {
			c, err := g.Coordinator()
			if err != nil {
				res.ErrorCode = protocol.CoordinatorNotAvailable
			} else {
				res.NodeId = int32(c.Id())
				res.Host = c.Host()
				res.Port = int32(c.Port())
			}
		}
	default:
		res.ErrorCode = protocol.Unknown
		res.ErrorMessage = fmt.Sprintf("unsupported key type %v in find coordinator request", r.KeyType)
		log.Errorf(res.ErrorMessage)
	}

	return rw.Write(res)
}
