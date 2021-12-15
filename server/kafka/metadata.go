package kafka

import (
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/metaData"
)

func (b *BrokerServer) metadata(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*metaData.Request)

	brokers := b.Cluster.Brokers()

	res := &metaData.Response{
		Brokers:   make([]metaData.ResponseBroker, 0, len(brokers)),
		Topics:    make([]metaData.ResponseTopic, 0, len(r.Topics)),
		ClusterId: "mokapi",
	}

	for _, b := range brokers {
		res.Brokers = append(res.Brokers, metaData.ResponseBroker{
			NodeId: int32(b.Id()),
			Host:   b.Host(),
			Port:   int32(b.Port()),
		})
	}

	var getTopic func(string) (Topic, protocol.ErrorCode)

	if len(r.Topics) > 0 {
		getTopic = func(name string) (Topic, protocol.ErrorCode) {
			if validateTopicName(name) != nil {
				return nil, protocol.InvalidTopic
			} else {
				topic := b.Cluster.Topic(name)
				if topic != nil {
					return topic, protocol.None
				} else {
					return nil, protocol.UnknownTopicOrPartition
				}
			}
		}
	} else {
		topics := make(map[string]Topic)
		for _, t := range b.Cluster.Topics() {
			topics[t.Name()] = t
			r.Topics = append(r.Topics, metaData.TopicName{Name: t.Name()})
		}
		getTopic = func(name string) (Topic, protocol.ErrorCode) {
			return topics[name], protocol.None
		}
	}

	for _, rt := range r.Topics {
		t, errCode := getTopic(rt.Name)
		if errCode != protocol.None {
			res.Topics = append(res.Topics, metaData.ResponseTopic{
				Name:      rt.Name,
				ErrorCode: errCode,
			})
			continue
		}

		resTopic := metaData.ResponseTopic{
			Name: t.Name(),
		}

		for i, p := range t.Partitions() {
			replicas := p.Replicas()
			nodes := make([]int32, 0, len(replicas))
			for _, n := range replicas {
				nodes = append(nodes, int32(n.Id()))
			}
			resTopic.Partitions = append(resTopic.Partitions, metaData.ResponsePartition{
				PartitionIndex: int32(i),
				LeaderId:       int32(p.Leader().Id()),
				ReplicaNodes:   nodes,
				IsrNodes:       nodes,
			})
		}

		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}
