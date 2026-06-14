package store

import (
	"mokapi/engine/common"
	"mokapi/kafka"
)

type Trigger func(topic string, partition int, record *kafka.Record, schemaId int) []*common.Action

type EventRecord struct {
	Api       string
	Topic     string
	Partition int
	Offset    int64
	Key       string
	Value     string
	SchemaId  int
	Headers   map[string]string
}
