package store

import (
	"encoding/json"
	"mokapi/kafka"
	"mokapi/runtime/events"
)

type LogRecord func(key, payload interface{}, headers []kafka.RecordHeader, partition int, offset int64, traits events.Traits)

type KafkaLog struct {
	Offset    int64             `json:"offset"`
	Key       string            `json:"key"`
	Message   string            `json:"message"`
	Partition int               `json:"partition"`
	Headers   map[string]string `json:"headers"`
}

func NewKafkaLog(key, payload interface{}, headers []kafka.RecordHeader, partition int, offset int64) *KafkaLog {
	log := &KafkaLog{
		Offset:    offset,
		Key:       toString(key),
		Message:   toString(payload),
		Partition: partition,
		Headers:   make(map[string]string),
	}
	for _, h := range headers {
		log.Headers[h.Key] = string(h.Value)
	}
	return log
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case kafka.Bytes:
		return kafka.BytesToString(val)
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}
