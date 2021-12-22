package kafka

import (
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/createTopics"
)

func (b *Broker) createtopics(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*createTopics.Request)
	res := &createTopics.Response{}

	for _, t := range r.Topics {
		_, err := b.Store.NewTopic(t.Name, int(t.NumPartitions))
		errCode := protocol.None
		if err != nil {
			errCode = protocol.TopicAlreadyExists
		}
		res.Topics = append(res.Topics, createTopics.TopicResponse{Name: t.Name, ErrorCode: errCode})
	}

	return rw.Write(res)
}
