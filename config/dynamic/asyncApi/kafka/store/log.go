package store

import (
	"mokapi/kafka"
	"mokapi/runtime/events"
	"time"
)

type LogRecord func(record kafka.Record, traits events.Traits)

type KafkaLog struct {
	Offset  int64
	Key     string
	Message string
	Time    time.Time
}

func NewKafkaLog(record kafka.Record) *KafkaLog {
	return &KafkaLog{
		Offset:  record.Offset,
		Key:     bytesToString(record.Key),
		Message: bytesToString(record.Value),
		Time:    record.Time,
	}
}
