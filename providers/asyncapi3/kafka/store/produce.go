package store

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/kafka/produce"
	"mokapi/runtime/monitor"
	"mokapi/schema/json/parser"
)

func (s *Store) produce(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*produce.Request)
	res := &produce.Response{}
	ctx := kafka.ClientFromContext(req)

	m, withMonitor := monitor.KafkaFromContext(req.Context)

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
				s := fmt.Sprintf("kafka: produce unknown topic %v", rt.Name)
				log.Errorf(s)
				resPartition.ErrorCode = kafka.UnknownTopicOrPartition
				resPartition.ErrorMessage = s
			} else if err := validateProducer(topic, ctx); err != nil {
				resPartition.ErrorCode = kafka.UnknownServerError
				resPartition.ErrorMessage = fmt.Sprintf("invalid producer clientId '%v' for topic %v: %v", ctx.ClientId, topic.Name, err)
				log.Errorf("kafka Produce: %v", resPartition.ErrorMessage)
			} else {
				p := topic.Partition(int(rp.Index))
				if p == nil {
					resPartition.ErrorCode = kafka.UnknownTopicOrPartition
					resPartition.ErrorMessage = fmt.Sprintf("unknown partition %v", rp.Index)
					log.Errorf("kafka Produce: %v", resPartition.ErrorMessage)
				} else {
					baseOffset, records, err := p.Write(rp.Record)
					if err != nil {
						resPartition.ErrorCode = kafka.InvalidRecord
						resPartition.ErrorMessage = fmt.Sprintf("invalid message received for topic %v: %v", rt.Name, err)
						resPartition.RecordErrors = records
						log.Errorf("kafka Produce: %v", resPartition.ErrorMessage)
					} else {
						resPartition.BaseOffset = baseOffset
						if withMonitor {
							go s.UpdateMetrics(m, topic, p, rp.Record)
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

func validateProducer(t *Topic, ctx *kafka.ClientContext) error {
	for _, op := range t.operations {
		if op.Action != "send" {
			continue
		}
		if op.Bindings.Kafka.ClientId != nil {
			s := op.Bindings.Kafka.ClientId
			p := parser.Parser{Schema: s}
			_, err := p.Parse(ctx.ClientId)
			return err
		}
	}

	return nil
}
