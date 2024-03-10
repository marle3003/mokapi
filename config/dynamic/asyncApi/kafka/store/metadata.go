package store

import (
	"mokapi/kafka"
	"mokapi/kafka/metaData"
)

func (s *Store) metadata(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*metaData.Request)

	brokers := s.Brokers()
	ctx := kafka.ClientFromContext(req)
	ctx.AllowAutoTopicCreation = r.AllowAutoTopicCreation

	res := &metaData.Response{
		Brokers:   make([]metaData.ResponseBroker, 0, len(brokers)),
		Topics:    make([]metaData.ResponseTopic, 0, len(r.Topics)),
		ClusterId: "mokapi",
	}

	for _, b := range brokers {
		res.Brokers = append(res.Brokers, metaData.ResponseBroker{
			NodeId: int32(b.Id),
			Host:   b.Host,
			Port:   int32(b.Port),
		})
	}

	b := s.getBrokerByHost(req.Host)
	var getTopic func(string) (*Topic, kafka.ErrorCode)

	if len(r.Topics) > 0 {
		getTopic = func(name string) (*Topic, kafka.ErrorCode) {
			if kafka.ValidateTopicName(name) != nil {
				return nil, kafka.InvalidTopic
			} else {
				topic := s.Topic(name)
				if topic != nil && isTopicAvailable(topic, b) {
					return topic, kafka.None
				} else {
					return nil, kafka.UnknownTopicOrPartition
				}
			}
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

		for i, p := range t.Partitions {
			replicas := p.Replicas
			nodes := make([]int32, 0, len(replicas))
			for _, n := range replicas {
				nodes = append(nodes, int32(n))
			}
			resTopic.Partitions = append(resTopic.Partitions, metaData.ResponsePartition{
				PartitionIndex: int32(i),
				LeaderId:       int32(p.Leader),
				ReplicaNodes:   nodes,
				IsrNodes:       nodes,
			})
		}

		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}

func isTopicAvailable(t *Topic, b *Broker) bool {
	if len(t.servers) == 0 {
		return true
	}
	for _, s := range t.servers {
		if s == b.Name {
			return true
		}
	}
	return false
}
