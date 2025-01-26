package store

import (
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
