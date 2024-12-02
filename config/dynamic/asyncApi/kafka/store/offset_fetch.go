package store

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/kafka/offsetFetch"
	"mokapi/schema/encoding"
	"mokapi/schema/json/schema"
)

func (s *Store) offsetFetch(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*offsetFetch.Request)
	res := &offsetFetch.Response{
		Topics: make([]offsetFetch.ResponseTopic, 0, len(r.Topics)),
	}

	ctx := kafka.ClientFromContext(req)

	for _, rt := range r.Topics {
		log.Infof("kafka OffsetFetch: topic %v, API Version=%v, client %v", rt.Name, req.Header.ApiVersion, ctx.ClientId)
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
				} else if _, ok := ctx.Member[r.GroupId]; !ok {
					log.Errorf("kafka OffsetFetch: unknown member topic=%v, client=%v", rt.Name, ctx.ClientId)
					resPartition.ErrorCode = kafka.UnknownMemberId
				} else {
					// todo check partition is assigned to member
					g, ok := s.Group(r.GroupId)
					if !ok {
						log.Errorf("kafka OffsetFetch: unkown group name %v, topic=%v, client=%v", r.GroupId, rt.Name, ctx.ClientId)
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

		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}

func validateConsumer(t *Topic, clientId, groupId string) (err error, code kafka.ErrorCode) {
	if t.Subscribe.ClientId != nil {
		_, err = encoding.Decode([]byte(clientId), encoding.WithSchema(&schema.Ref{Value: t.Publish.ClientId}))
		if err != nil {
			return fmt.Errorf("invalid clientId: %v", err), kafka.UnknownServerError
		}
	}
	if t.Subscribe.GroupId != nil {
		_, err = encoding.Decode([]byte(groupId), encoding.WithSchema(&schema.Ref{Value: t.Publish.GroupId}))
		if err != nil {
			return fmt.Errorf("invalid groupId: %v", err), kafka.InvalidGroupId
		}
	}
	return
}
