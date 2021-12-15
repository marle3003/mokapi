package kafka

import (
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/offset"
)

func (b *BrokerServer) offset(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*offset.Request)
	res := &offset.Response{Topics: make([]offset.ResponseTopic, 0)}

	for _, rt := range r.Topics {
		topic := b.Cluster.Topic(rt.Name)

		resPartitions := make([]offset.ResponsePartition, 0)
		for _, rp := range rt.Partitions {
			resPartition := offset.ResponsePartition{
				Index:     rp.Index,
				Timestamp: rp.Timestamp,
			}
			if topic == nil {
				resPartition.ErrorCode = protocol.UnknownTopicOrPartition
			} else {
				partition := topic.Partition(int(rp.Index))
				if partition == nil {
					resPartition.ErrorCode = protocol.UnknownTopicOrPartition
				} else {
					switch {
					case rp.Timestamp == protocol.Earliest || rp.Timestamp == 0:
						resPartition.Offset = partition.StartOffset()
					case rp.Timestamp == protocol.Latest:
						resPartition.Offset = partition.Offset()
					default:
						// TODO
						// look up the offsets for the given partitions by timestamp. The returned offset
						// for each partition is the earliest offset for which the timestamp is greater
						// than or equal to the given timestamp.
					}
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
