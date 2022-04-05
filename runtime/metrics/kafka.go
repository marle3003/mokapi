package metrics

type Kafka struct {
	Messages *Counter
}

func NewKafka() *Kafka {
	return &Kafka{
		Messages: NewCounter("kafka.messages.total"),
	}
}
