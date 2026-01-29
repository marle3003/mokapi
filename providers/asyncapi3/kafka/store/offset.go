package store

import (
	"mokapi/kafka"
	"mokapi/kafka/offset"

	log "github.com/sirupsen/logrus"
)

func (s *Store) offset(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*offset.Request)
	res := &offset.Response{Topics: make([]offset.ResponseTopic, 0)}

	ctx := kafka.ClientFromContext(req.Context)

	reqLog := &KafkaListOffsetsRequest{
		Topics: make(map[string][]KafkaListOffsetsRequestPartition),
	}
	resLog := &KafkaListOffsetsResponse{
		Topics: make(map[string][]KafkaListOffsetsResponsePartition),
	}

	for _, rt := range r.Topics {
		topic := s.Topic(rt.Name)

		resPartitions := make([]offset.ResponsePartition, 0)
		for _, rp := range rt.Partitions {
			resPartition := offset.ResponsePartition{
				Index:     rp.Index,
				Timestamp: rp.Timestamp,
			}

			reqLog.Topics[rt.Name] = append(reqLog.Topics[rt.Name], KafkaListOffsetsRequestPartition{
				Partition: int(rp.Index),
			})

			if topic == nil {
				log.Errorf("kafka Offset: unknown topic %v, client=%v", topic, ctx.ClientId)
				resPartition.ErrorCode = kafka.UnknownTopicOrPartition
			} else {
				partition := topic.Partition(int(rp.Index))
				if partition == nil {
					log.Errorf("kafka Offset: unknown partition %v, topic=%v, client=%v", rp.Index, rt.Name, ctx.ClientId)
					resPartition.ErrorCode = kafka.UnknownTopicOrPartition
				} else {
					if req.Header.ApiVersion > 0 {
						switch {
						case rp.Timestamp == kafka.Earliest || rp.Timestamp == 0:
							resPartition.Offset = partition.StartOffset()
							resPartition.Timestamp = partition.OffsetTimestamp(resPartition.Offset)
						case rp.Timestamp == kafka.Latest:
							resPartition.Offset = partition.Offset()
							resPartition.Timestamp = partition.OffsetTimestamp(resPartition.Offset)
						default:
							// TODO
							// look up the offsets for the given partitions by timestamp. The returned offset
							// for each partition is the earliest offset for which the timestamp is greater
							// than or equal to the given timestamp.
							log.Errorf("kafka Offset: only supporting timestamp=latest|earliest")
							resPartition.ErrorCode = kafka.UnknownServerError
						}
					} else {
						if rp.Timestamp == kafka.Earliest && rp.MaxNumOffsets == 1 {
							resPartition.OldStyleOffsets = append(resPartition.OldStyleOffsets, partition.StartOffset())
						} else if rp.Timestamp == kafka.Latest && rp.MaxNumOffsets == 1 {
							resPartition.OldStyleOffsets = append(resPartition.OldStyleOffsets, partition.Offset())
						} else {
							log.Errorf("kafka Offset: only supporting timestamp=latest|earliest and max_num_offsets=1")
							resPartition.ErrorCode = kafka.UnknownServerError
						}
					}
					resLog.Topics[rt.Name] = append(resLog.Topics[rt.Name], KafkaListOffsetsResponsePartition{
						Partition: int(rp.Index),
						Timestamp: resPartition.Timestamp,
						Snapshot: KafkaListOffsetsResponseSnapshot{
							StartOffset: partition.StartOffset(),
							EndOffset:   partition.Offset(),
						},
					})
				}
			}
			resPartitions = append(resPartitions, resPartition)

		}

		res.Topics = append(res.Topics, offset.ResponseTopic{
			Name:       rt.Name,
			Partitions: resPartitions,
		})
	}

	go func() {
		l := &KafkaRequestLogEvent{
			Request:  reqLog,
			Response: resLog,
		}
		s.logRequest(req.Header)(l)
	}()

	return rw.Write(res)
}
