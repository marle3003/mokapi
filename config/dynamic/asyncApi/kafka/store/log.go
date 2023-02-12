package store

import (
	"mokapi/kafka"
	"mokapi/runtime/events"
)

type LogRecord func(record kafka.Record, partition int, traits events.Traits)

type KafkaLog struct {
	Offset    int64             `json:"offset"`
	Key       string            `json:"key"`
	Message   string            `json:"message"`
	Partition int               `json:"partition"`
	Headers   map[string]string `json:"headers"`
}

func NewKafkaLog(record kafka.Record, partition int) *KafkaLog {
	log := &KafkaLog{
		Offset:    record.Offset,
		Key:       kafka.BytesToString(record.Key),
		Message:   kafka.BytesToString(record.Value),
		Partition: partition,
		Headers:   make(map[string]string),
	}
	for _, h := range record.Headers {
		log.Headers[h.Key] = string(h.Value)
	}
	return log
}
