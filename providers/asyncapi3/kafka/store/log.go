package store

import (
	"fmt"
	"mokapi/kafka"
	"mokapi/runtime/events"
	"strings"
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

func (s *Store) logRequest(h *kafka.Header) func(log *KafkaRequestLogEvent) {
	return func(log *KafkaRequestLogEvent) {
		log.Api = s.cluster
		log.Header.set(h)
		t := events.NewTraits().
			WithNamespace("kafka").
			WithName(s.cluster).
			With("type", "request").
			With("clientId", h.ClientId)
		_ = s.eh.Push(log, t)
	}
}

type KafkaRequestLogEvent struct {
	Api      string             `json:"api"`
	Header   KafkaRequestHeader `json:"header"`
	Request  KafkaRequest       `json:"request"`
	Response any                `json:"response"`
}

func (l *KafkaRequestLogEvent) Title() string {
	return l.Request.Title()
}

type KafkaRequest interface {
	Title() string
}

type KafkaRequestHeader struct {
	RequestKey  kafka.ApiKey `json:"requestKey"`
	RequestName string       `json:"requestName"`
	Version     int16        `json:"version"`
}

func (h *KafkaRequestHeader) set(header *kafka.Header) {
	h.RequestKey = header.ApiKey
	h.RequestName = strings.Split(header.ApiKey.String(), " ")[0]
	h.Version = header.ApiVersion
}

type KafkaResponseError struct {
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type KafkaJoinGroupRequest struct {
	GroupName    string   `json:"groupName"`
	MemberId     string   `json:"memberId"`
	ProtocolType string   `json:"protocolType"`
	Protocols    []string `json:"protocols"`
}

func (r *KafkaJoinGroupRequest) Title() string {
	return fmt.Sprintf("JoinGroup %s", r.GroupName)
}

type KafkaJoinGroupResponse struct {
	GenerationId int32    `json:"generationId"`
	ProtocolName string   `json:"protocolName"`
	MemberId     string   `json:"memberId"`
	LeaderId     string   `json:"leaderId"`
	Members      []string `json:"members,omitempty"`
}

type KafkaSyncGroupRequest struct {
	GroupName        string                              `json:"groupName"`
	GenerationId     int32                               `json:"generationId"`
	MemberId         string                              `json:"memberId"`
	ProtocolType     string                              `json:"protocolType"`
	ProtocolName     string                              `json:"protocolName"`
	GroupAssignments map[string]KafkaSyncGroupAssignment `json:"groupAssignments,omitempty"`
}

type KafkaSyncGroupAssignment struct {
	Version int16            `json:"version"`
	Topics  map[string][]int `json:"topics"`
}

func (r *KafkaSyncGroupRequest) Title() string {
	return fmt.Sprintf("SyncGroup %s", r.GroupName)
}

type KafkaSyncGroupResponse struct {
	ProtocolType string                   `json:"protocolType"`
	ProtocolName string                   `json:"protocolName"`
	Assignment   KafkaSyncGroupAssignment `json:"assignment"`
}

type KafkaListOffsetsRequest struct {
	Topics map[string][]KafkaListOffsetsRequestPartition `json:"topics"`
}

func (r *KafkaListOffsetsRequest) Title() string {
	return "ListOffsets"
}

type KafkaListOffsetsRequestPartition struct {
	Partition int   `json:"partition"`
	Timestamp int64 `json:"timestamp"`
}

type KafkaListOffsetsResponse struct {
	Topics map[string][]KafkaListOffsetsResponsePartition `json:"topics"`
}

type KafkaListOffsetsResponsePartition struct {
	Partition int                              `json:"partition"`
	Timestamp int64                            `json:"timestamp"`
	Offset    int64                            `json:"offset"`
	Snapshot  KafkaListOffsetsResponseSnapshot `json:"snapshot"`
}

type KafkaListOffsetsResponseSnapshot struct {
	StartOffset int64 `json:"startOffset"`
	EndOffset   int64 `json:"endOffset"`
}

type KafkaFindCoordinatorRequest struct {
	Key     string `json:"key"`
	KeyType int8   `json:"keyType"`
}

func (r *KafkaFindCoordinatorRequest) Title() string {
	return "FindCoordinator"
}

type KafkaFindCoordinatorResponse struct {
	KafkaResponseError
	Host string `json:"host"`
	Port int    `json:"port"`
}

type KafkaInitProducerIdRequest struct {
	TransactionalId      string `json:"transactionalId"`
	TransactionTimeoutMs int32  `json:"transactionTimeoutMs"`
	ProducerId           int64  `json:"producerId"`
	ProducerEpoch        int16  `json:"producerEpoch"`
	Enable2PC            bool   `json:"enable2PC"`
}

func (r *KafkaInitProducerIdRequest) Title() string {
	return "InitProducerId"
}

type KafkaInitProducerIdResponse struct {
	KafkaResponseError
	ProducerId              int64 `json:"producerId"`
	ProducerEpoch           int16 `json:"producerEpoch"`
	OngoingTxnProducerId    int64 `json:"ongoingTxnProducerId"`
	OngoingTxnProducerEpoch int16 `json:"ongoingTxnProducerEpoch"`
}
