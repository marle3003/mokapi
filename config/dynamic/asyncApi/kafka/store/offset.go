package store

import (
	"mokapi/kafka"
	"mokapi/kafka/offset"
)

func (s *Store) offset(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*offset.Request)
	res := &offset.Response{Topics: make([]offset.ResponseTopic, 0)}

	for _, rt := range r.Topics {
		topic := s.Topic(rt.Name)

		resPartitions := make([]offset.ResponsePartition, 0)
		for _, rp := range rt.Partitions {
			resPartition := offset.ResponsePartition{
				Index:     rp.Index,
				Timestamp: rp.Timestamp,
			}
			if topic == nil {
				resPartition.ErrorCode = kafka.UnknownTopicOrPartition
			} else {
				partition := topic.Partition(int(rp.Index))
				if partition == nil {
					resPartition.ErrorCode = kafka.UnknownTopicOrPartition
				} else {
					switch {
					case rp.Timestamp == kafka.Earliest || rp.Timestamp == 0:
						resPartition.Offset = partition.StartOffset()
					case rp.Timestamp == kafka.Latest:
						resPartition.Offset = partition.Offset()
					default:
						// TODO
						// look up the offsets for the given partitions by timestamp. The returned offset
						// for each partition is the earliest offset for which the timestamp is greater
						// than or equal to the given timestamp.
					}
					resPartition.OldStyleOffsets = resPartition.Offset
				}
			}
			resPartitions = append(resPartitions, resPartition)
		}

		res.Topics = append(res.Topics, offset.ResponseTopic{
			Name:       rt.Name,
			Partitions: resPartitions,
		})
	}

	return rw.Write(res)
}
