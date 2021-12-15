package kafka

import (
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/produce"
)

func (b *BrokerServer) produce(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*produce.Request)
	res := &produce.Response{}

	for _, rt := range r.Topics {
		topic := b.Cluster.Topic(rt.Name)
		resTopic := produce.ResponseTopic{
			Name: rt.Name,
		}

		if topic == nil {
			resTopic.ErrorCode = protocol.UnknownTopicOrPartition
		} else {
			p := topic.Partition(int(rt.Data.Partition))
			if p == nil {
				resTopic.ErrorCode = protocol.UnknownTopicOrPartition
			} else {
				p.Write(rt.Data.Record)
				resTopic.StartOffset = p.StartOffset()
				resTopic.Offset = p.Offset()
			}
		}

		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}
