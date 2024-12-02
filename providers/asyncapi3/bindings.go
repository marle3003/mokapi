package asyncapi3

type ServerBindings struct {
	Kafka BrokerBindings `yaml:"kafka" json:"kafka"`
}

type ChannelBindings struct {
	Kafka TopicBindings `yaml:"kafka" json:"kafka"`
}

type OperationBindings struct {
	Kafka KafkaOperation `yaml:"kafka" json:"kafka"`
}

type MessageBinding struct {
	Kafka KafkaMessageBinding `yaml:"kafka" json:"kafka"`
}
