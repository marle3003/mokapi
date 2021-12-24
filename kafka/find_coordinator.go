package kafka

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/findCoordinator"
)

func (b *Broker) findCoordinator(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*findCoordinator.Request)
	res := &findCoordinator.Response{}

	switch r.KeyType {
	case findCoordinator.KeyTypeGroup:
		g := b.Store.GetOrCreateGroup(r.Key, b.Id)
		c := g.Coordinator()
		if c == nil {
			res.ErrorCode = protocol.CoordinatorNotAvailable
		} else {
			res.NodeId = int32(c.Id())
			res.Host = c.Host()
			res.Port = int32(c.Port())
		}

	default:
		res.ErrorCode = protocol.Unknown
		res.ErrorMessage = fmt.Sprintf("unsupported key type %v in find coordinator request", r.KeyType)
		log.Errorf(res.ErrorMessage)
	}

	return rw.Write(res)
}
