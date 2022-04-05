package store

import (
	"mokapi/kafka"
	"mokapi/kafka/offsetFetch"
)

func (s *Store) offsetFetch(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*offsetFetch.Request)
	res := &offsetFetch.Response{
		Topics: make([]offsetFetch.ResponseTopic, 0, len(r.Topics)),
	}

	ctx := kafka.ClientFromContext(req)

	for _, rt := range r.Topics {
		topic := s.Topic(rt.Name)
		resTopic := offsetFetch.ResponseTopic{Name: rt.Name, Partitions: make([]offsetFetch.Partition, 0, len(rt.PartitionIndexes))}

		for _, index := range rt.PartitionIndexes {
			resPartition := &offsetFetch.Partition{Index: index, CommittedOffset: -1}

			if topic == nil {
				resPartition.ErrorCode = kafka.UnknownTopicOrPartition
			} else {
				p := topic.Partition(int(index))
				if p == nil {
					resPartition.ErrorCode = kafka.UnknownTopicOrPartition
				} else if _, ok := ctx.Member[r.GroupId]; !ok {
					resPartition.ErrorCode = kafka.UnknownMemberId
				} else {
					// todo check partition is assigned to member
					g, ok := s.Group(r.GroupId)
					if !ok {
						resPartition.ErrorCode = kafka.InvalidGroupId
					} else {
						resPartition.CommittedOffset = g.Offset(topic.Name, p.Index)
					}
				}
			}

			if req.Header.ApiVersion == 0 && resPartition.CommittedOffset == -1 {
				resPartition.ErrorCode = kafka.UnknownTopicOrPartition
			}

			resTopic.Partitions = append(resTopic.Partitions, *resPartition)
		}

		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}
