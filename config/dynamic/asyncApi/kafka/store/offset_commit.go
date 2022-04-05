package store

import (
	"mokapi/kafka"
	"mokapi/kafka/offsetCommit"
)

func (s *Store) offsetCommit(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*offsetCommit.Request)
	res := &offsetCommit.Response{
		Topics: make([]offsetCommit.ResponseTopic, 0, len(r.Topics)),
	}

	ctx := kafka.ClientFromContext(req)

	for _, rt := range r.Topics {
		topic := s.Topic(rt.Name)
		resTopic := offsetCommit.ResponseTopic{
			Name:       rt.Name,
			Partitions: make([]offsetCommit.ResponsePartition, 0, len(rt.Partitions)),
		}
		for _, rp := range rt.Partitions {
			resPartition := offsetCommit.ResponsePartition{
				Index: rp.Index,
			}

			if topic == nil {
				resPartition.ErrorCode = kafka.UnknownTopicOrPartition
			} else {
				p := topic.Partition(int(rp.Index))
				if p == nil {
					resPartition.ErrorCode = kafka.UnknownTopicOrPartition
				} else if _, ok := ctx.Member[r.GroupId]; !ok {
					resPartition.ErrorCode = kafka.UnknownMemberId
				} else {
					// todo check partition is assigned to member
					if rp.Offset > p.Offset() {
						resPartition.ErrorCode = kafka.OffsetOutOfRange
					} else {
						g, ok := s.Group(r.GroupId)
						if !ok {
							resPartition.ErrorCode = kafka.InvalidGroupId
						} else {
							g.Commit(topic.Name, p.Index, rp.Offset)
						}
					}
				}
			}

			resTopic.Partitions = append(resTopic.Partitions, resPartition)
		}
		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}
