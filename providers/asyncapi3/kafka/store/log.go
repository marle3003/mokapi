package store

import (
	"encoding/json"
	"mokapi/kafka"
	"mokapi/runtime/events"
)

type LogRecord func(log *KafkaLog, traits events.Traits)

type KafkaLog struct {
	Offset    int64             `json:"offset"`
	Key       LogValue          `json:"key"`
	Message   LogValue          `json:"message"`
	SchemaId  int               `json:"schemaId"`
	MessageId string            `json:"messageId"`
	Partition int               `json:"partition"`
	Headers   map[string]string `json:"headers"`
}

type LogValue struct {
	Value  string `json:"value"`
	Binary []byte `json:"binary"`
}

func NewKafkaLog(record *kafka.Record, key, payload interface{}, schemaId int, partition int) *KafkaLog {
	log := &KafkaLog{
		Offset:    record.Offset,
		Key:       toValue(record.Key, key),
		Message:   toValue(record.Value, payload),
		SchemaId:  schemaId,
		Partition: partition,
		Headers:   make(map[string]string),
	}
	for _, h := range record.Headers {
		log.Headers[h.Key] = string(h.Value)
	}
	return log
}

func toValue(b kafka.Bytes, v interface{}) LogValue {
	lv := LogValue{
		Binary: kafka.Read(b),
	}

	switch val := v.(type) {
	case string:
		lv.Value = val
	default:
		b, _ := json.Marshal(val)
		lv.Value = string(b)
	}

	return lv
}
