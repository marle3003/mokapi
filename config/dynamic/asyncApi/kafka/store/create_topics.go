package store

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/kafka"
	"mokapi/kafka/createTopics"
)

func (s *Store) createtopics(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*createTopics.Request)
	res := &createTopics.Response{}

	for _, t := range r.Topics {
		op := &asyncApi.Operation{}
		config := &asyncApi.Channel{
			Subscribe: op,
			Publish:   op,
		}
		config.Bindings.Kafka.Partitions = int(t.NumPartitions)

		_, err := s.NewTopic(t.Name, config)
		errCode := kafka.None
		if err != nil {
			errCode = kafka.TopicAlreadyExists
		}
		res.Topics = append(res.Topics, createTopics.TopicResponse{Name: t.Name, ErrorCode: errCode})
	}

	return rw.Write(res)
}
