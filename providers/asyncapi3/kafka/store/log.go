package store

import (
	"mokapi/runtime/events"
)

type LogRecord func(log *KafkaLog, traits events.Traits)

type KafkaLog struct {
	Offset    int64               `json:"offset"`
	Key       LogValue            `json:"key"`
	Message   LogValue            `json:"message"`
	SchemaId  int                 `json:"schemaId"`
	MessageId string              `json:"messageId"`
	Partition int                 `json:"partition"`
	Headers   map[string]LogValue `json:"headers"`
	Deleted   bool                `json:"deleted"`
	Api       string              `json:"api"`
}

type LogValue struct {
	Value  string `json:"value"`
	Binary []byte `json:"binary"`
}

func (l *KafkaLog) Title() string {
	if l.Key.Value != "" {
		return l.Key.Value
	} else {
		return string(l.Key.Binary)
	}
}
