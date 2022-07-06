package store

import (
	"mokapi/kafka"
	"mokapi/runtime/events"
)

type LogRecord func(record kafka.Record, traits events.Traits)

type KafkaLog struct {
	Offset  int64  `json:"offset"`
	Key     string `json:"key"`
	Message string `json:"message"`
}

func NewKafkaLog(record kafka.Record) *KafkaLog {
	return &KafkaLog{
		Offset:  record.Offset,
		Key:     kafka.BytesToString(record.Key),
		Message: kafka.BytesToString(record.Value),
	}
}
