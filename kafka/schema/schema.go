package schema

type Cluster struct {
	Topics  []Topic
	Brokers []Broker
}

func New() Cluster {
	return Cluster{}
}

type Topic struct {
	Name       string
	Partitions []Partition
}

type Partition struct {
	Index    int
	Replicas []int
}

type Broker struct {
	Id   int
	Host string
	Port int
}
