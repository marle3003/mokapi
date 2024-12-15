package store

import (
	"mokapi/kafka"
	"mokapi/kafka/createTopics"
	"mokapi/providers/asyncapi3"
)

func (s *Store) createtopics(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*createTopics.Request)
	res := &createTopics.Response{}

	for _, t := range r.Topics {
		channel := &asyncapi3.Channel{
			Title: t.Name,
		}
		channel.Bindings.Kafka.Partitions = int(t.NumPartitions)
		ops := []*asyncapi3.Operation{
			{
				Action:  "send",
				Channel: asyncapi3.ChannelRef{Value: channel},
			},
			{
				Action:  "receive",
				Channel: asyncapi3.ChannelRef{Value: channel},
			},
		}

		_, err := s.NewTopic(t.Name, channel, ops)
		errCode := kafka.None
		if err != nil {
			errCode = kafka.TopicAlreadyExists
		}
		res.Topics = append(res.Topics, createTopics.TopicResponse{Name: t.Name, ErrorCode: errCode})
	}

	return rw.Write(res)
}
