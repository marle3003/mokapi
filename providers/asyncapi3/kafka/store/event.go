package store

import (
	"mokapi/kafka"
)

type Trigger func(record *kafka.Record, schemaId int) bool

type EventRecord struct {
	Offset   int64
	Key      string
	Value    string
	SchemaId int
	Headers  map[string]string
}
