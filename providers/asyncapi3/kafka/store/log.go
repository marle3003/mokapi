package store

import (
	"mokapi/kafka"
	"mokapi/runtime/events"
)

type LogRecord func(log *KafkaLog, traits events.Traits)

type KafkaLog struct {
	Offset         int64               `json:"offset"`
	Key            LogValue            `json:"key"`
	Message        LogValue            `json:"message"`
	SchemaId       int                 `json:"schemaId"`
	MessageId      string              `json:"messageId"`
	Partition      int                 `json:"partition"`
	Headers        map[string]LogValue `json:"headers"`
	ProducerId     int64               `json:"producerId"`
	ProducerEpoch  int16               `json:"producerEpoch"`
	SequenceNumber int32               `json:"sequenceNumber"`
	Deleted        bool                `json:"deleted"`
	Api            string              `json:"api"`
	ClientId       string              `json:"clientId"`
	ScriptFile     string              `json:"script"`
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

func newKafkaLog(record *kafka.Record) *KafkaLog {
	return &KafkaLog{
		Key:            LogValue{Binary: kafka.Read(record.Key)},
		Message:        LogValue{Binary: kafka.Read(record.Value)},
		Headers:        convertHeader(record.Headers),
		ProducerId:     record.ProducerId,
		ProducerEpoch:  record.ProducerEpoch,
		SequenceNumber: record.SequenceNumber,
	}
}
