package store

import (
	"fmt"
	"mokapi/kafka"
	"mokapi/runtime/events"
)

type LogRecord func(log *KafkaMessageLog, traits events.Traits)

type KafkaMessageLog struct {
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

func (l *KafkaMessageLog) Title() string {
	if l.Key.Value != "" {
		return l.Key.Value
	} else {
		return string(l.Key.Binary)
	}
}

func newKafkaLog(record *kafka.Record) *KafkaMessageLog {
	return &KafkaMessageLog{
		Key:            LogValue{Binary: kafka.Read(record.Key)},
		Message:        LogValue{Binary: kafka.Read(record.Value)},
		Headers:        convertHeader(record.Headers),
		ProducerId:     record.ProducerId,
		ProducerEpoch:  record.ProducerEpoch,
		SequenceNumber: record.SequenceNumber,
	}
}

type KafkaRequestData interface {
	Title() string
}

type KafkaRequestLog struct {
	Api     string           `json:"api"`
	Request KafkaRequestData `json:"request"`
}

type KafkaRequestBase struct {
	RequestKey  kafka.ApiKey `json:"requestKey"`
	RequestName string       `json:"requestName"`
}

type KafkaJoinGroupRequest struct {
	KafkaRequestBase
	GroupName    string   `json:"groupName"`
	MemberId     string   `json:"memberId"`
	ProtocolType string   `json:"protocolType"`
	Protocols    []string `json:"protocols"`
}

func (l *KafkaRequestLog) Title() string {
	return l.Request.Title()
}

func (r *KafkaJoinGroupRequest) Title() string {
	return fmt.Sprintf("JoinGroup %s", r.GroupName)
}
