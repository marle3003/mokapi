package kafka

import (
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/offsetFetch"
)

func (b *BrokerServer) offsetFetch(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*offsetFetch.Request)
	res := &offsetFetch.Response{
		Topics: make([]offsetFetch.ResponseTopic, 0, len(r.Topics)),
	}

	ctx := getClientContext(req)

	for _, rt := range r.Topics {
		topic := b.Cluster.Topic(rt.Name)
		resTopic := offsetFetch.ResponseTopic{Name: rt.Name, Partitions: make([]offsetFetch.Partition, 0, len(rt.PartitionIndexes))}

		for _, index := range rt.PartitionIndexes {
			resPartition := &offsetFetch.Partition{Index: index, CommittedOffset: -1}

			if topic == nil {
				resPartition.ErrorCode = protocol.UnknownTopicOrPartition
			} else {
				p := topic.Partition(int(index))
				if p == nil {
					resPartition.ErrorCode = protocol.UnknownTopicOrPartition
				} else if _, ok := ctx.member[r.GroupId]; !ok {
					resPartition.ErrorCode = protocol.UnknownMemberId
				} else {
					// todo check partition is assigned to member
					g := b.Cluster.Group(r.GroupId)
					resPartition.CommittedOffset = g.Offset(topic.Name(), p.Index())
				}
			}

			if req.Header.ApiVersion == 0 && resPartition.CommittedOffset == -1 {
				resPartition.ErrorCode = protocol.UnknownTopicOrPartition
			}

			resTopic.Partitions = append(resTopic.Partitions, *resPartition)
		}

		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}
