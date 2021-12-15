package kafka

import (
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/offsetCommit"
)

func (b *BrokerServer) offsetCommit(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*offsetCommit.Request)
	res := &offsetCommit.Response{
		Topics: make([]offsetCommit.ResponseTopic, 0, len(r.Topics)),
	}

	ctx := getClientContext(req)

	for _, rt := range r.Topics {
		topic := b.Cluster.Topic(rt.Name)
		resTopic := offsetCommit.ResponseTopic{
			Name:       rt.Name,
			Partitions: make([]offsetCommit.ResponsePartition, 0, len(rt.Partitions)),
		}
		for _, rp := range rt.Partitions {
			resPartition := offsetCommit.ResponsePartition{
				Index: rp.Index,
			}

			if topic == nil {
				resPartition.ErrorCode = protocol.UnknownTopicOrPartition
			} else {
				p := topic.Partition(int(rp.Index))
				if p == nil {
					resPartition.ErrorCode = protocol.UnknownTopicOrPartition
				} else if _, ok := ctx.member[r.GroupId]; !ok {
					resPartition.ErrorCode = protocol.UnknownMemberId
				} else {
					// todo check partition is assigned to member
					if rp.Offset > p.Offset() {
						resPartition.ErrorCode = protocol.OffsetOutOfRange
					} else {
						g := b.Cluster.Group(r.GroupId)
						g.Commit(topic.Name(), p.Index(), rp.Offset)
					}
				}
			}

			resTopic.Partitions = append(resTopic.Partitions, resPartition)
		}
		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}
