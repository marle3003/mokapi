package store

import (
	log "github.com/sirupsen/logrus"
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
		log.Infof("offsetFetch topic %v, API Version=%v, client %v", rt.Name, req.Header.ApiVersion, ctx.ClientId)
		topic := s.Topic(rt.Name)
		resTopic := offsetFetch.ResponseTopic{Name: rt.Name, Partitions: make([]offsetFetch.Partition, 0, len(rt.PartitionIndexes))}

		for _, index := range rt.PartitionIndexes {
			resPartition := &offsetFetch.Partition{Index: index, CommittedOffset: -1}

			if topic == nil {
				log.Errorf("kafka: offsetCommit unknown topic %v, client=%v", topic, ctx.ClientId)
				resPartition.ErrorCode = kafka.UnknownTopicOrPartition
			} else {
				p := topic.Partition(int(index))
				if p == nil {
					log.Errorf("kafka: offsetCommit unknown partition %v, topic=%v, client=%v", index, rt.Name, ctx.ClientId)
					resPartition.ErrorCode = kafka.UnknownTopicOrPartition
				} else if _, ok := ctx.Member[r.GroupId]; !ok {
					log.Errorf("kafka: offsetCommit unknown member topic=%v, client=%v", rt.Name, ctx.ClientId)
					resPartition.ErrorCode = kafka.UnknownMemberId
				} else {
					// todo check partition is assigned to member
					g, ok := s.Group(r.GroupId)
					if !ok {
						log.Errorf("kafka: invalid group name %v, topic=%v, client=%v", r.GroupId, rt.Name, ctx.ClientId)
						resPartition.ErrorCode = kafka.InvalidGroupId
					} else {
						resPartition.CommittedOffset = g.Offset(topic.Name, p.Index)
						log.Infof("kafka: offsetFetch committed offset %v, topic=%v, partition=%v, client=%v", resPartition.CommittedOffset, rt.Name, index, ctx.ClientId)
					}
				}
			}

			resTopic.Partitions = append(resTopic.Partitions, *resPartition)
		}

		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}
