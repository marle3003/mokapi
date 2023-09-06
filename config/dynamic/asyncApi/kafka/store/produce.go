package store

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/kafka/produce"
	"mokapi/runtime/monitor"
)

func (s *Store) produce(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*produce.Request)
	res := &produce.Response{}

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
				log.Errorf("kafka: produce unknown topic %v", rt.Name)
				resPartition.ErrorCode = kafka.UnknownTopicOrPartition
			} else {
				p := topic.Partition(int(rp.Index))
				if p == nil {
					log.Errorf("kafka: produce unknown partition %v", rp.Index)
					resPartition.ErrorCode = kafka.UnknownTopicOrPartition
				} else {
					baseOffset, err := p.Write(rp.Record)
					if err != nil {
						s := fmt.Sprintf("kafka: invalid message received for topic %v: %v", rt.Name, err)
						log.Errorf(s)
						resPartition.ErrorCode = kafka.CorruptMessage
						resPartition.ErrorMessage = s
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
