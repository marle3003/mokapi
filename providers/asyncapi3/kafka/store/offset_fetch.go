package store

import (
	"fmt"
	"mokapi/kafka"
	"mokapi/kafka/offsetFetch"
	"mokapi/schema/json/parser"

	log "github.com/sirupsen/logrus"
)

func (s *Store) offsetFetch(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*offsetFetch.Request)
	res := &offsetFetch.Response{}

	ctx := kafka.ClientFromContext(req.Context)

	if req.Header.ApiVersion >= 8 {
		for _, g := range r.Groups {
			res.Groups = append(res.Groups, offsetFetch.ResponseGroup{
				GroupId: g.GroupId,
				Topics:  s.fetchTopicOffsets(g.GroupId, g.Topics, ctx),
			})
		}
	} else {
		res.Topics = s.fetchTopicOffsets(r.GroupId, r.Topics, ctx)
	}

	return rw.Write(res)
}

func (s *Store) fetchTopicOffsets(groupId string, topics []offsetFetch.RequestTopic, ctx *kafka.ClientContext) []offsetFetch.ResponseTopic {
	result := make([]offsetFetch.ResponseTopic, 0, len(topics))
	for _, rt := range topics {
		topic := s.Topic(rt.Name)
		resTopic := offsetFetch.ResponseTopic{Name: rt.Name, Partitions: make([]offsetFetch.Partition, 0, len(rt.PartitionIndexes))}

		for _, index := range rt.PartitionIndexes {
			resPartition := &offsetFetch.Partition{Index: index, CommittedOffset: -1}

			if topic == nil {
				log.Errorf("kafka OffsetFetch: unknown topic %v, client=%v", rt.Name, ctx.ClientId)
				resPartition.ErrorCode = kafka.UnknownTopicOrPartition
			} else {
				p := topic.Partition(int(index))
				if p == nil {
					log.Errorf("kafka OffsetFetch: unknown partition %v, topic=%v, client=%v", index, rt.Name, ctx.ClientId)
					resPartition.ErrorCode = kafka.UnknownTopicOrPartition
				} else if _, ok := ctx.Member[groupId]; !ok {
					log.Errorf("kafka OffsetFetch: unknown member topic=%v, client=%v", rt.Name, ctx.ClientId)
					resPartition.ErrorCode = kafka.UnknownMemberId
				} else {
					// todo check partition is assigned to member
					g, ok := s.Group(groupId)
					if !ok {
						log.Errorf("kafka OffsetFetch: unkown group name %v, topic=%v, client=%v", groupId, rt.Name, ctx.ClientId)
						resPartition.ErrorCode = kafka.GroupIdNotFound
					} else {
						if err, code := validateConsumer(topic, ctx.ClientId, g.Name); err != nil {
							log.Errorf("kafka OffsetFetch: invalid consumer '%v' for topic %v: %v", ctx.ClientId, rt.Name, err)
							resPartition.ErrorCode = code
						} else {
							resPartition.CommittedOffset = g.Offset(topic.Name, p.Index)
						}
					}
				}
			}

			resTopic.Partitions = append(resTopic.Partitions, *resPartition)
		}

		result = append(result, resTopic)
	}
	return result
}

func validateConsumer(t *Topic, clientId, groupId string) (error, kafka.ErrorCode) {
	var last error
	var code kafka.ErrorCode
	var err error
	for _, op := range t.operations {
		if op.Action != "receive" {
			continue
		}
		if op.Bindings.Kafka.ClientId != nil {
			p := parser.Parser{Schema: op.Bindings.Kafka.ClientId}
			_, err = p.Parse(clientId)
			if err != nil {
				last = fmt.Errorf("invalid clientId: %v", err)
				code = kafka.UnknownServerError
				continue
			}
		}
		if op.Bindings.Kafka.GroupId != nil {
			p := parser.Parser{Schema: op.Bindings.Kafka.GroupId}
			_, err = p.Parse(groupId)
			if err != nil {
				last = fmt.Errorf("invalid groupId: %v", err)
				code = kafka.InvalidGroupId
				continue
			}
		}
		return nil, kafka.None
	}

	return last, code
}
