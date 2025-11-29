package store

import (
	"fmt"
	"mokapi/kafka"
	"mokapi/kafka/produce"
	"mokapi/runtime/monitor"
	"mokapi/schema/json/parser"

	log "github.com/sirupsen/logrus"
)

func (s *Store) produce(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*produce.Request)
	res := &produce.Response{}
	ctx := kafka.ClientFromContext(req)

	m, withMonitor := monitor.KafkaFromContext(req.Context)
	opts := WriteOptions{}

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
				s := fmt.Sprintf("kafka: failed to write: %s", rt.Name)
				log.Error(s)
				resPartition.ErrorCode = kafka.UnknownTopicOrPartition
				resPartition.ErrorMessage = s
			} else if err := validateProducer(topic, ctx); err != nil {
				resPartition.ErrorCode = kafka.UnknownServerError
				resPartition.ErrorMessage = fmt.Sprintf("invalid producer clientId '%v' for topic %v: %v", ctx.ClientId, topic.Name, err)
				log.Errorf("kafka: failed to write to topic '%s': %s", topic.Name, resPartition.ErrorMessage)
			} else {
				p := topic.Partition(int(rp.Index))
				if p == nil {
					resPartition.ErrorCode = kafka.UnknownTopicOrPartition
					resPartition.ErrorMessage = fmt.Sprintf("unknown partition %v", rp.Index)
					log.Errorf("kafka: failed to write to topic '%s': %s", topic.Name, resPartition.ErrorMessage)
				} else {
					wr, err := p.write(rp.Record, opts)
					if err != nil {
						resPartition.ErrorCode = kafka.UnknownServerError
						resPartition.ErrorMessage = fmt.Sprintf("failed to write to topic '%v': %v", rt.Name, err.Error())
						log.Errorf("kafka: failed to write to topic '%s' partition %d: %s", topic.Name, rp.Index, resPartition.ErrorMessage)
					} else if wr.ErrorCode != kafka.None {
						resPartition.ErrorCode = wr.ErrorCode
						resPartition.ErrorMessage = wr.ErrorMessage
						resPartition.RecordErrors = wr.Records
						log.Errorf("kafka: failed to write to topic '%s' partition %d: %s", topic.Name, rp.Index, resPartition.ErrorMessage)
					} else {
						resPartition.BaseOffset = wr.BaseOffset
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
