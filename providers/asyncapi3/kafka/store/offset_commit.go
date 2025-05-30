package store

import (
	"context"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/kafka/offsetCommit"
	"mokapi/runtime/monitor"
	"strconv"
)

func (s *Store) offsetCommit(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*offsetCommit.Request)
	res := &offsetCommit.Response{
		Topics: make([]offsetCommit.ResponseTopic, 0, len(r.Topics)),
	}

	ctx := kafka.ClientFromContext(req)

	for _, rt := range r.Topics {
		log.Infof("kafa OffsetCommit: topic %v, client=%v", rt.Name, ctx.ClientId)
		topic := s.Topic(rt.Name)
		if topic == nil {
			log.Errorf("kafka OffsetCommit: unknown topic %v, client=%v", rt.Name, ctx.ClientId)
		}

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
					log.Errorf("kafka OffsetCommit: unknown partition %v, topic=%v, client=%v", rp.Index, topic, ctx.ClientId)
					resPartition.ErrorCode = kafka.UnknownTopicOrPartition
				} else if _, ok := ctx.Member[r.GroupId]; !ok {
					log.Errorf("kafka OffsetCommit: unknown member topic=%v, client=%v", topic.Name, ctx.ClientId)
					resPartition.ErrorCode = kafka.UnknownMemberId
				} else {
					// todo check partition is assigned to member
					if rp.Offset > p.Offset() {
						log.Errorf("kafka OffsetCommit: offset out of range, offset=%v, topic=%v, client=%v", rp.Offset, topic, ctx.ClientId)
						resPartition.ErrorCode = kafka.OffsetOutOfRange
					} else {
						g, ok := s.Group(r.GroupId)
						if !ok {
							log.Errorf("kafka OffsetCommit: unknown group name %v, topic=%v, client=%v", r.GroupId, topic, ctx.ClientId)
							resPartition.ErrorCode = kafka.GroupIdNotFound
						} else {
							if err, code := validateConsumer(topic, ctx.ClientId, g.Name); err != nil {
								log.Errorf("kafka OffsetCommit: invalid consumer '%v' for topic %v: %v", ctx.ClientId, rt.Name, err)
								resPartition.ErrorCode = code
							} else {
								g.Commit(topic.Name, p.Index, rp.Offset)
								go s.processMetricsOffsetCommit(req.Context, g, topic.Name, p)
							}
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

func (s *Store) processMetricsOffsetCommit(ctx context.Context, g *Group, topic string, partition *Partition) {
	m, ok := monitor.KafkaFromContext(ctx)
	if !ok {
		return
	}

	lag := float64(partition.Offset() - g.Commits[topic][partition.Index])
	m.Lags.WithLabel(s.cluster, g.Name, topic, strconv.Itoa(partition.Index)).Set(lag)
}
