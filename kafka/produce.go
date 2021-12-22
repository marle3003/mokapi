package kafka

import (
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/produce"
)

func (b *Broker) produce(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*produce.Request)
	res := &produce.Response{}

	for _, rt := range r.Topics {
		topic := b.Store.Topic(rt.Name)
		resTopic := produce.ResponseTopic{
			Name: rt.Name,
		}

		for _, rp := range rt.Partitions {
			resPartition := produce.ResponsePartition{
				Index: rp.Index,
			}

			if topic == nil {
				resPartition.ErrorCode = protocol.UnknownTopicOrPartition
			} else {
				p := topic.Partition(int(rp.Index))
				if p == nil {
					resPartition.ErrorCode = protocol.UnknownTopicOrPartition
				} else {
					baseOffset := p.Write(rp.Record)
					resPartition.BaseOffset = baseOffset
				}
			}

			resTopic.Partitions = append(resTopic.Partitions, resPartition)
		}

		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}
