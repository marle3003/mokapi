package memory

type Schema struct {
	Topics  []TopicSchema
	Brokers []BrokerSchema
}

type TopicSchema struct {
	Name       string
	Partitions []PartitionSchema
}

type PartitionSchema struct {
	Index    int
	Replicas []int
}

type BrokerSchema struct {
	Id   int
	Host string
	Port int
}

func NewCluster(schema Schema) *Cluster {
	c := &Cluster{
		topics:  make(map[string]*Topic),
		brokers: make(map[int]*Broker),
		groups:  make(map[string]*Group),
	}
	for _, b := range schema.Brokers {
		c.brokers[b.Id] = &Broker{
			id:   b.Id,
			host: b.Host,
			port: b.Port,
		}
	}
	for _, ts := range schema.Topics {
		t, _ := c.addTopic(ts.Name)
		for _, p := range ts.Partitions {
			replicas := make([]*Broker, 0)
			for _, id := range p.Replicas {
				replicas = append(replicas, c.brokers[id])
			}

			part := newPartition(p.Index, replicas)
			t.partitions[p.Index] = part

		}
	}
	return c
}
