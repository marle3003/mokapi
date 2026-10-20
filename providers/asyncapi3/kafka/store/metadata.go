package store

import (
	"mokapi/kafka"
	"mokapi/kafka/metaData"
	"mokapi/providers/asyncapi3"
	"path"
)

func (s *Store) metadata(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*metaData.Request)

	brokers := s.Brokers()

	res := &metaData.Response{
		Brokers:   make([]metaData.ResponseBroker, 0, len(brokers)),
		Topics:    make([]metaData.ResponseTopic, 0, len(r.Topics)),
		ClusterId: s.config.Info.Name,
	}

	// Mokapi does no leader management; therefore, only the current server is returned as the broker.
	host, port := parseHostAndPort(req.Host)
	b := s.getBrokerByPort(req.Host)
	if b != nil && b.Host != "" {
		host = b.Host
	}
	res.Brokers = append(res.Brokers, metaData.ResponseBroker{
		NodeId: 0,
		Host:   host,
		Port:   int32(port),
	})

	var getTopic func(string) (*Topic, kafka.ErrorCode)

	if len(r.Topics) > 0 {
		getTopic = func(name string) (*Topic, kafka.ErrorCode) {
			if kafka.ValidateTopicName(name) != nil {
				return nil, kafka.InvalidTopic
			}
			topic := s.Topic(name)
			if topic != nil && isTopicAvailable(topic, b) {
				return topic, kafka.None
			}
			if topic == nil && r.AllowAutoTopicCreation {
				var err error
				topic, err = s.NewTopic(name, b)
				if err == nil {
					return topic, kafka.None
				}
			}
			return nil, kafka.UnknownTopicOrPartition
		}
	} else {
		topics := make(map[string]*Topic)
		for _, t := range s.Topics() {
			if !isTopicAvailable(t, b) {
				continue
			}
			topics[t.Name] = t
			r.Topics = append(r.Topics, metaData.TopicName{Name: t.Name})
		}
		getTopic = func(name string) (*Topic, kafka.ErrorCode) {
			return topics[name], kafka.None
		}
	}

	for _, rt := range r.Topics {
		t, errCode := getTopic(rt.Name)
		if errCode != kafka.None {
			res.Topics = append(res.Topics, metaData.ResponseTopic{
				Name:      rt.Name,
				ErrorCode: errCode,
			})
			continue
		}

		resTopic := metaData.ResponseTopic{
			Name: t.Name,
		}

		for i := range t.Partitions {
			resTopic.Partitions = append(resTopic.Partitions, metaData.ResponsePartition{
				PartitionIndex: int32(i),
				LeaderId:       0,
			})
		}

		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}

func isTopicAvailable(t *Topic, b *Broker) bool {
	return isChannelAvailable(t.Config, b)
}

func isChannelAvailable(ch *asyncapi3.Channel, b *Broker) bool {
	if len(ch.Servers) == 0 {
		return true
	}
	for _, s := range ch.Servers {
		name := path.Base(s.Ref)
		if b != nil && name == b.Name {
			return true
		}
	}
	return false
}
