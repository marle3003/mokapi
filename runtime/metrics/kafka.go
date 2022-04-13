package metrics

type Kafka struct {
	Messages *CounterMap
}

func NewKafka() *Kafka {
	return &Kafka{
		Messages: NewCounterMap("kafka_messages_total"),
	}
}
