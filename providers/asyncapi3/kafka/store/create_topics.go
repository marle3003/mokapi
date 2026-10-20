package store

import (
	"mokapi/kafka"
	"mokapi/kafka/createTopics"
)

func (s *Store) createTopics(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*createTopics.Request)
	res := &createTopics.Response{}
	b := s.getBrokerByPort(req.Host)

	for _, rt := range r.Topics {
		t := s.Topic(rt.Name)
		if t != nil {
			appendCreateTopicResponse(res, createTopics.TopicResponse{
				Name:      rt.Name,
				ErrorCode: kafka.TopicAlreadyExists,
			})
			continue
		}

		t, err := s.NewTopic(rt.Name, b)
		if err != nil {
			appendCreateTopicResponse(res, createTopics.TopicResponse{
				Name:         rt.Name,
				ErrorCode:    kafka.PolicyViolation,
				ErrorMessage: err.Error(),
			})
			continue
		}

		if rt.NumPartitions >= 0 {
			t.updatePartitions(int(rt.NumPartitions))
		}

		appendCreateTopicResponse(res, createTopics.TopicResponse{
			Name:          rt.Name,
			NumPartitions: int32(len(t.Partitions)),
		})
	}

	return rw.Write(res)
}

func appendCreateTopicResponse(res *createTopics.Response, topic createTopics.TopicResponse) {
	res.Topics = append(res.Topics, topic)
}
