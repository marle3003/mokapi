package common

type KafkaEventRecord struct {
	Api       string
	Topic     string
	Partition int
	Offset    int64
	Key       string
	Value     string
	SchemaId  int
	Headers   map[string]string
}

type KafkaFilter struct{}
