package store

import (
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/kafka/produce"
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
					}
				}
			}

			resTopic.Partitions = append(resTopic.Partitions, resPartition)
		}

		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}
