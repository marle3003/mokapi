package store

import (
	"context"
	log "github.com/sirupsen/logrus"
	"io"
	"mokapi/kafka"
	"mokapi/kafka/produce"
	"mokapi/runtime/monitor"
	"strconv"
	"time"
)

func (s *Store) produce(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*produce.Request)
	res := &produce.Response{}

	for _, rt := range r.Topics {
		topic := s.Topic(rt.Name)
		resTopic := produce.ResponseTopic{
			Name: rt.Name,
		}

		for _, rp := range rt.Partitions {
			resPartition := produce.ResponsePartition{
				Index: rp.Index,
			}

			if topic == nil {
				resPartition.ErrorCode = kafka.UnknownTopicOrPartition
			} else {
				p := topic.Partition(int(rp.Index))
				if p == nil {
					resPartition.ErrorCode = kafka.UnknownTopicOrPartition
				} else {
					baseOffset, err := p.Write(rp.Record)
					if err != nil {
						log.Infof("kafka corrupt message: %v", err)
						resPartition.ErrorCode = kafka.CorruptMessage
					} else {
						resPartition.BaseOffset = baseOffset
						go s.processMetricsProduce(req.Context, rt.Name, p, rp.Record)
					}
				}
			}

			resTopic.Partitions = append(resTopic.Partitions, resPartition)
		}

		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}

func (s *Store) processMetricsProduce(ctx context.Context, topic string, partition *Partition, records kafka.RecordBatch) {
	m, ok := monitor.KafkaFromContext(ctx)
	if !ok {
		return
	}

	m.Messages.WithLabel(s.cluster, topic).Add(1)
	m.LastMessage.WithLabel(s.cluster, topic).Set(float64(time.Now().Unix()))

	for name, g := range s.groups {
		gt, ok := g.Commits[topic]
		if !ok {
			continue
		}
		commit, ok := gt[partition.Index]
		if !ok {
			continue
		}
		lag := float64(partition.Offset() - commit)
		m.Lags.WithLabel(s.cluster, name, topic, strconv.Itoa(partition.Index)).Set(lag)
	}
}

func bytesToString(bytes kafka.Bytes) string {
	bytes.Seek(0, io.SeekStart)
	b := make([]byte, bytes.Len())
	bytes.Read(b)
	return string(b)
}
